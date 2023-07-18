package dbauth

import (
	"bytes"
	"errors"
	"github.com/codingbeard/cbweb/cbwebauth"
	"github.com/codingbeard/checkmail"
	"github.com/golang-jwt/jwt"
	"github.com/jinzhu/gorm"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type Provider struct {
	db                   GormReadWrite
	log                  Logger
	secret               string
	getUserRecordsFunc   func() []UserRecord
	generateAuthHashFunc func(user UserRecord) string
	hashWorkFactor       int
	saveUserRecordFunc   func(user UserRecord) error
}

type GormReadWrite interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type Logger interface {
	InfoF(category string, message string, args ...interface{})
}

type UserRecord interface {
	GetEmail() string
	SetEmail(email string)
	GetPassword() string
	SetPassword(password string)
	GetCreated() time.Time
	GetPermissions() []string
}

type Dependencies struct {
	Db                   GormReadWrite
	Log                  Logger
	Secret               string
	GetUserRecordsFunc   func() []UserRecord
	SaveUserRecordFunc   func(user UserRecord) error
	GenerateAuthHashFunc func(user UserRecord) string
	HashWorkFactor       int
}

type UserClaim struct {
	Email    string `json:"email"`
	AuthHash string `json:"auth_hash"`
	jwt.StandardClaims
}

var ProviderName = "dbauth"
var cookieKey = "cbauth"

func New(dependencies Dependencies) (*Provider, error) {
	if dependencies.GetUserRecordsFunc == nil {
		return nil, errors.New("missing GetUserRecordsFunc")
	}
	if dependencies.GenerateAuthHashFunc == nil {
		return nil, errors.New("missing GenerateAuthHashFunc")
	}
	auth := &Provider{
		db:                   dependencies.Db,
		log:                  dependencies.Log,
		secret:               dependencies.Secret,
		getUserRecordsFunc:   dependencies.GetUserRecordsFunc,
		generateAuthHashFunc: dependencies.GenerateAuthHashFunc,
		hashWorkFactor:       dependencies.HashWorkFactor,
		saveUserRecordFunc:   dependencies.SaveUserRecordFunc,
	}

	return auth, nil
}

func (a *Provider) getUserFromLoginToken(tokenString string) UserRecord {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	})
	if err != nil {
		return nil
	}

	if token.Valid {
		userClaim, ok := token.Claims.(*UserClaim)
		if ok {
			for _, user := range a.getUserRecordsFunc() {
				if strings.ToLower(user.GetEmail()) == strings.ToLower(userClaim.Email) {
					authHash := a.generateAuthHashFunc(user)
					if authHash == userClaim.AuthHash {
						return user
					}
				}
			}
		}
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return nil
		} else {
			return nil
		}
	} else {
		return nil
	}

	return nil
}

func (a *Provider) GenerateLoginToken(user UserRecord) (string, error) {
	claims := UserClaim{
		Email:    user.GetEmail(),
		AuthHash: a.generateAuthHashFunc(user),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
			Issuer:    "dbauth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString([]byte(a.secret))
}

func (a *Provider) GetProviderName() string {
	return ProviderName
}

func (a *Provider) GetUniqueIdentifier(ctx *fasthttp.RequestCtx) string {
	if ctx.Request.URI().QueryArgs().Has("dbauthtoken") {
		user := a.getUserFromLoginToken(string(ctx.Request.URI().QueryArgs().Peek("dbauthtoken")))
		if user != nil {
			return user.GetEmail()
		}
	}

	cookie := ctx.Request.Header.Cookie(cookieKey)
	if len(cookie) == 0 {
		return ""
	}

	user := a.getUserFromLoginToken(string(cookie))
	if user == nil {
		return ""
	}

	return user.GetEmail()
}

func (a *Provider) GetPermissions(ctx *fasthttp.RequestCtx) []string {
	if ctx.Request.URI().QueryArgs().Has("dbauthtoken") {
		user := a.getUserFromLoginToken(string(ctx.Request.URI().QueryArgs().Peek("dbauthtoken")))
		if user != nil {
			return append(user.GetPermissions(), cbwebauth.LoggedIn)
		}
	}

	cookie := ctx.Request.Header.Cookie(cookieKey)
	if len(cookie) == 0 {
		return []string{}
	}

	user := a.getUserFromLoginToken(string(cookie))
	if user == nil {
		return []string{}
	}

	return append(user.GetPermissions(), cbwebauth.LoggedIn)
}

func (a *Provider) IsAuthenticated(ctx *fasthttp.RequestCtx) bool {
	if ctx.Request.URI().QueryArgs().Has("dbauthtoken") {
		return a.getUserFromLoginToken(string(ctx.Request.URI().QueryArgs().Peek("dbauthtoken"))) != nil
	}

	cookie := ctx.Request.Header.Cookie(cookieKey)
	if len(cookie) == 0 {
		return false
	}

	return a.getUserFromLoginToken(string(cookie)) != nil
}

func (a *Provider) Login(ctx *fasthttp.RequestCtx) (bool, map[string]error) {
	post := ctx.Request.PostArgs()
	if post != nil && post.Len() > 0 {
		validationErrors := make(map[string]error)

		if !post.Has("email") {
			validationErrors["email"] = errors.New("please provide an email")
		}

		if !post.Has("password") {
			validationErrors["password"] = errors.New("please provide a password")
		}

		if checkmail.ValidateFormat(string(post.Peek("email"))) != nil {
			validationErrors["email"] = errors.New("please provide a valid email")
		}

		if len(validationErrors) > 0 {
			return false, validationErrors
		}

		for _, user := range a.getUserRecordsFunc() {
			if strings.ToLower(user.GetEmail()) == strings.ToLower(string(post.Peek("email"))) {
				if bcrypt.CompareHashAndPassword([]byte(user.GetPassword()), post.Peek("password")) == nil {
					e := a.SetAuthCookie(ctx, user)
					if e != nil {
						return false, map[string]error{"flash": errors.New("error setting auth cookie")}
					}
					return true, validationErrors
				} else {
					return false, map[string]error{"password": errors.New("invalid password")}
				}
			}
		}
		return false, map[string]error{"email": errors.New("user not found")}
	} else if ctx.Request.URI().QueryArgs().Has("dbauthtoken") {
		user := a.getUserFromLoginToken(string(ctx.Request.URI().QueryArgs().Peek("dbauthtoken")))
		if user != nil {
			e := a.SetAuthCookie(ctx, user)
			if e != nil {
				return false, map[string]error{"flash": errors.New("error setting auth cookie")}
			}
			return true, make(map[string]error)
		}
		return false, map[string]error{"email": errors.New("user not found")}
	}

	return false, map[string]error{"flash": errors.New("invalid request")}
}

func (a *Provider) SetAuthCookie(ctx *fasthttp.RequestCtx, user UserRecord) error {
	token, e := a.GenerateLoginToken(user)
	if e != nil {
		return e
	}
	var cookie fasthttp.Cookie
	cookie.SetExpire(time.Now().Add(time.Hour * 24 * 365))
	cookie.SetHTTPOnly(true)
	cookie.SetPath("")
	cookie.SetKey(cookieKey)
	cookie.SetValue(token)
	ctx.Response.Header.SetCookie(&cookie)

	return nil
}

func (a *Provider) Logout(ctx *fasthttp.RequestCtx) bool {
	var cookie fasthttp.Cookie
	cookie.SetExpire(time.Now().Add(-time.Hour))
	cookie.SetHTTPOnly(true)
	cookie.SetPath("")
	cookie.SetKey(cookieKey)
	ctx.Response.Header.SetCookie(&cookie)

	return true
}

func (a *Provider) Register(ctx *fasthttp.RequestCtx) (bool, map[string]error) {
	post := ctx.Request.PostArgs()
	if post == nil {
		return false, map[string]error{"flash": errors.New("invalid request")}
	}

	validationErrors := make(map[string]error)

	if !post.Has("email") {
		validationErrors["email"] = errors.New("please provide an email")
	}

	if !post.Has("password") {
		validationErrors["password"] = errors.New("please provide a password")
	}

	if !post.Has("password-confirm") {
		validationErrors["password-confirm"] = errors.New("please confirm your password")
	}

	if checkmail.ValidateFormat(string(post.Peek("email"))) != nil {
		validationErrors["email"] = errors.New("please provide a valid email")
	}

	if bytes.Compare(post.Peek("password"), post.Peek("password-confirm")) != 0 {
		validationErrors["password-confirm"] = errors.New("please make sure your password confirmation matches your password")
	}

	if len(validationErrors) > 0 {
		return false, validationErrors
	}

	e := checkmail.ValidateHost(string(post.Peek("email")))
	if errors.Is(e, checkmail.ErrUnresolvableHost) {
		validationErrors["email"] = errors.New("please provide a real email account")
	}

	if len(validationErrors) > 0 {
		return false, validationErrors
	}

	//todo complete registration

	return true, nil
}

func (a *Provider) ChangePassword(ctx *fasthttp.RequestCtx) (bool, map[string]error) {
	post := ctx.Request.PostArgs()
	if post == nil {
		return false, map[string]error{"flash": errors.New("invalid request")}
	}

	validationErrors := make(map[string]error)

	if !post.Has("email") {
		validationErrors["email"] = errors.New("please provide an email")
	}

	if !post.Has("password") {
		validationErrors["password"] = errors.New("please provide a password")
	}

	if !post.Has("password-confirm") {
		validationErrors["password-confirm"] = errors.New("please confirm your password")
	}

	if checkmail.ValidateFormat(string(post.Peek("email"))) != nil {
		validationErrors["email"] = errors.New("please provide a valid email")
	}

	if bytes.Compare(post.Peek("password"), post.Peek("password-confirm")) != 0 {
		validationErrors["password-confirm"] = errors.New("please make sure your password confirmation matches your password")
	}

	if len(validationErrors) > 0 {
		return false, validationErrors
	}

	e := checkmail.ValidateHost(string(post.Peek("email")))
	if errors.Is(e, checkmail.ErrUnresolvableHost) {
		validationErrors["email"] = errors.New("please provide a real email account")
		return false, validationErrors
	}

	var user UserRecord
	for _, potentialUser := range a.getUserRecordsFunc() {
		if strings.ToLower(potentialUser.GetEmail()) == string(post.Peek("email")) {
			user = potentialUser
		}
	}

	if user == nil {
		validationErrors["email"] = errors.New("that user does not exist")
		return false, validationErrors
	}

	password, e := bcrypt.GenerateFromPassword(post.Peek("password"), a.hashWorkFactor)
	if e != nil {
		validationErrors["flash"] = errors.New("there was an error changing your password")
		return false, validationErrors
	}

	user.SetPassword(string(password))
	e = a.saveUserRecordFunc(user)
	if e != nil {
		validationErrors["flash"] = errors.New("there was an error saving your updated password")
		return false, validationErrors
	}

	return true, nil
}

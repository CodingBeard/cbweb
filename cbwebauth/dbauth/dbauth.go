package dbauth

import (
	"bytes"
	"errors"
	"github.com/codingbeard/checkmail"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type Provider struct {
	lastCache             int64
	cache                 []UserRecord
	db                    GormReadWrite
	log                   Logger
	secret                string
	reloadUserRecordsFunc func() []UserRecord
	generateAuthHashFunc  func(user UserRecord) string
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
	GetPassword() string
	GetCreated() time.Time
}

type Dependencies struct {
	Db                    GormReadWrite
	Log                   Logger
	Secret                string
	ReloadUserRecordsFunc func() []UserRecord
	GenerateAuthHashFunc  func(user UserRecord) string
}

type UserClaim struct {
	Email    string `json:"email"`
	AuthHash string `json:"auth_hash"`
	jwt.StandardClaims
}

var ProviderName = "dbauth"
var cookieKey = "cbauth"

func New(dependencies Dependencies) (*Provider, error) {
	if dependencies.ReloadUserRecordsFunc == nil {
		return nil, errors.New("missing ReloadUserRecordsFunc")
	}
	if dependencies.GenerateAuthHashFunc == nil {
		return nil, errors.New("missing GenerateAuthHashFunc")
	}
	auth := &Provider{
		db:                    dependencies.Db,
		log:                   dependencies.Log,
		secret:                dependencies.Secret,
		reloadUserRecordsFunc: dependencies.ReloadUserRecordsFunc,
		generateAuthHashFunc:  dependencies.GenerateAuthHashFunc,
	}

	return auth, nil
}

func (a *Provider) reloadUsers() {
	now := time.Now().Unix()
	if a.lastCache < now-1 {
		if a.reloadUserRecordsFunc != nil {
			a.cache = a.reloadUserRecordsFunc()
		}
		a.lastCache = now
	}
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
			for _, user := range a.cache {
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

func (a *Provider) generateLoginToken(user UserRecord) (string, error) {
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
	a.reloadUsers()
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

func (a *Provider) IsAuthenticated(ctx *fasthttp.RequestCtx) bool {
	a.reloadUsers()
	cookie := ctx.Request.Header.Cookie(cookieKey)
	if len(cookie) == 0 {
		return false
	}

	return a.getUserFromLoginToken(string(cookie)) != nil
}

func (a *Provider) Login(ctx *fasthttp.RequestCtx) (bool, map[string]error) {
	a.reloadUsers()
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

		for _, user := range a.cache {
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
	token, e := a.generateLoginToken(user)
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
	a.reloadUsers()
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

	if checkmail.ValidateHost(string(post.Peek("email"))) != nil {
		validationErrors["email"] = errors.New("please provide a real email account")
	}

	if len(validationErrors) > 0 {
		return false, validationErrors
	}

	//todo complete registration

	return true, nil
}

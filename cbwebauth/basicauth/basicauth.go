package basicauth

import (
	"bytes"
	"encoding/base64"
	"github.com/codingbeard/cbweb/cbwebauth"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
)

type Credential struct {
	Username string
	Password string
	Permissions []string
}

type Provider struct {
	Credentials []Credential
}

var ProviderName = "basicauth"

func New(credentials ...Credential) *Provider {
	return &Provider{Credentials: credentials}
}

func (p Provider) GetProviderName() string {
	return ProviderName
}

func (p Provider) GetUniqueIdentifier(ctx *fasthttp.RequestCtx) string {
	user, _ := p.getCredentials(ctx)

	return user
}

func (p Provider) GetPermissions(ctx *fasthttp.RequestCtx) []string {
	user, _ := p.getCredentials(ctx)

	if user != "" {
		for _, credential := range p.Credentials {
			if credential.Username == user {
				return append(credential.Permissions, cbwebauth.LoggedIn)
			}
		}
	}

	return []string{}
}

func (p Provider) IsAuthenticated(ctx *fasthttp.RequestCtx) bool {
	user, pass := p.getCredentials(ctx)
	for _, credential := range p.Credentials {
		if credential.Username == user && bcrypt.CompareHashAndPassword([]byte(credential.Password), []byte(pass)) == nil {
			return true
		}
	}

	ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)

	return false
}

func (p Provider) Login(ctx *fasthttp.RequestCtx) (bool, map[string]error) {
	return true, make(map[string]error)
}

func (p Provider) Logout(ctx *fasthttp.RequestCtx) bool {
	ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)

	return false
}

func (p Provider) Register(ctx *fasthttp.RequestCtx) (bool, map[string]error) {
	return true, make(map[string]error)
}

func (p Provider) ChangePassword(ctx *fasthttp.RequestCtx) (bool, map[string]error) {
	return true, make(map[string]error)
}

func (p Provider) getCredentials(ctx *fasthttp.RequestCtx) (string, string) {
	auth := ctx.Request.Header.Peek("Authorization")
	if bytes.HasPrefix(auth, []byte("Basic ")) {
		payload, err := base64.StdEncoding.DecodeString(string(auth[len([]byte("Basic ")):]))
		if err == nil {
			pair := bytes.SplitN(payload, []byte(":"), 2)
			if len(pair) == 2 {
				return string(pair[0]), string(pair[1])
			}
		}
	}

	return "", ""
}

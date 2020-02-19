package cbwebauth

import (
	"errors"
	"github.com/valyala/fasthttp"
)

type Provider interface {
	GetProviderName() string
	GetUniqueIdentifier(ctx *fasthttp.RequestCtx) string
	IsAuthenticated(ctx *fasthttp.RequestCtx) bool
	Login(ctx *fasthttp.RequestCtx) (bool, []error)
	Logout(ctx *fasthttp.RequestCtx) bool
	Register(ctx *fasthttp.RequestCtx) (bool, []error)
}

type Container struct {
	providers               []Provider
	unauthorisedRedirectUri string
	logoutRedirectUri       string
}

type Config struct {
	Providers               []Provider
	UnauthorisedRedirectUri string
	LogoutRedirectUri       string
}

func New(config Config) *Container {
	container := &Container{
		providers:               config.Providers,
		unauthorisedRedirectUri: config.UnauthorisedRedirectUri,
		logoutRedirectUri:       config.LogoutRedirectUri,
	}

	return container
}

func (c *Container) AuthMiddleware(ctx *fasthttp.RequestCtx) (bool, error) {
	if len(c.providers) == 0 {
		return true, nil
	}

	for _, provider := range c.providers {
		if provider.IsAuthenticated(ctx) {
			return true, nil
		} else {
			ok, _ := provider.Login(ctx)
			if ok {
				return true, nil
			}
		}
	}

	if c.unauthorisedRedirectUri != "" {
		ctx.Redirect(c.unauthorisedRedirectUri, 302)
	}

	return false, nil
}

func (c *Container) GetUniqueIdentifier(providerName string, ctx *fasthttp.RequestCtx) (string, error) {
	if len(c.providers) == 0 {
		return "", errors.New("no auth providers configured")
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			return provider.GetUniqueIdentifier(ctx), nil
		}
	}

	return "", errors.New("auth provider not found")
}

func (c *Container) IsAuthenticated(providerName string, ctx *fasthttp.RequestCtx) (bool, error) {
	if len(c.providers) == 0 {
		return true, errors.New("no auth providers configured")
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			return provider.IsAuthenticated(ctx), nil
		}
	}

	return false, errors.New("auth provider not found")
}

func (c *Container) Login(providerName string, ctx *fasthttp.RequestCtx) (bool, error, []error) {
	if len(c.providers) == 0 {
		return true, errors.New("no auth providers configured"), []error{}
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			loginSuccess, userErrors := provider.Login(ctx)
			return loginSuccess, nil, userErrors
		}
	}

	return false, errors.New("auth provider not found"), []error{}
}

func (c *Container) Register(providerName string, ctx *fasthttp.RequestCtx) (bool, error, []error) {
	if len(c.providers) == 0 {
		return true, errors.New("no auth providers configured"), []error{}
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			registerSuccess, userErrors := provider.Register(ctx)
			return registerSuccess, nil, userErrors
		}
	}

	return false, errors.New("auth provider not found"), []error{}
}

func (c *Container) Logout(providerName string, ctx *fasthttp.RequestCtx) (bool, error) {
	if len(c.providers) == 0 {
		return true, errors.New("no auth providers configured")
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			ok := provider.Logout(ctx)

			if c.logoutRedirectUri != "" {
				ctx.Redirect(c.logoutRedirectUri, 302)
			}

			return ok, nil
		}
	}

	return false, errors.New("auth provider not found")
}

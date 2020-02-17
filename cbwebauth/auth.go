package cbwebauth

import (
	"errors"
	"github.com/valyala/fasthttp"
)

type Provider interface {
	GetProviderName() string
	GetUniqueIdentifier(ctx *fasthttp.RequestCtx) string
	IsAuthenticated(ctx *fasthttp.RequestCtx) bool
	Login(ctx *fasthttp.RequestCtx) bool
	Logout(ctx *fasthttp.RequestCtx) bool
	Register(ctx *fasthttp.RequestCtx) bool
}

type Container struct {
	providers []Provider
}

func New(providers ...Provider) *Container {
	container := &Container{
		providers: providers,
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
		}
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

func (c *Container) Login(providerName string, ctx *fasthttp.RequestCtx) (bool, error) {
	if len(c.providers) == 0 {
		return true, errors.New("no auth providers configured")
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			return provider.Login(ctx), nil
		}
	}

	return false, errors.New("auth provider not found")
}

func (c *Container) Register(providerName string, ctx *fasthttp.RequestCtx) (bool, error) {
	if len(c.providers) == 0 {
		return true, errors.New("no auth providers configured")
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			return provider.Register(ctx), nil
		}
	}

	return false, errors.New("auth provider not found")
}

func (c *Container) Logout(providerName string, ctx *fasthttp.RequestCtx) (bool, error) {
	if len(c.providers) == 0 {
		return true, errors.New("no auth providers configured")
	}

	for _, provider := range c.providers {
		if provider.GetProviderName() == providerName {
			return provider.Logout(ctx), nil
		}
	}

	return false, errors.New("auth provider not found")
}
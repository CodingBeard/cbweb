package cbweb

import (
	"github.com/valyala/fasthttp"
	"webshield/state"
)

type MiddlewareHandler struct {
	middleware []func(ctx *fasthttp.RequestCtx) (bool, error)
	final      func(ctx *fasthttp.RequestCtx)
}

func (m MiddlewareHandler) AddMiddleware(middleware ...func(ctx *fasthttp.RequestCtx) (bool, error)) MiddlewareHandler {
	m.middleware = middleware

	return m
}

func (m MiddlewareHandler) SetFinal(final func(ctx *fasthttp.RequestCtx)) MiddlewareHandler {
	m.final = final

	return m
}

func (m MiddlewareHandler) Handle(ctx *fasthttp.RequestCtx) {
	for _, middleware := range m.middleware {
		ok, e := middleware(ctx)
		if e != nil {
			state.Errors.Error(e)
		}
		if !ok {
			return
		}
	}

	if m.final != nil {
		m.final(ctx)
	}
}

func HtmlMiddleware(ctx *fasthttp.RequestCtx) (bool, error) {
	ctx.SetContentType("text/html")

	return true, nil
}
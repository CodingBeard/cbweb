package cbweb

import (
	"github.com/didip/tollbooth/config"
	"github.com/didip/tollbooth_fasthttp"
	"github.com/valyala/fasthttp"
)

type Profiler interface {
	Start()
	End()
}

type MiddlewareHandler struct {
	ErrorHandler ErrorHandler
	Limiter      *config.Limiter
	middleware   []func(ctx *fasthttp.RequestCtx) (bool, error)
	final        func(ctx *fasthttp.RequestCtx)
	afterFinal   []func(ctx *fasthttp.RequestCtx) (bool, error)
	profiler     Profiler
}

func (m MiddlewareHandler) AddMiddleware(middleware ...func(ctx *fasthttp.RequestCtx) (bool, error)) MiddlewareHandler {
	m.middleware = append(m.middleware, middleware...)

	return m
}

func (m MiddlewareHandler) SetFinal(final func(ctx *fasthttp.RequestCtx)) MiddlewareHandler {
	m.final = final

	return m
}

func (m MiddlewareHandler) SetAfterFinal(after ...func(ctx *fasthttp.RequestCtx) (bool, error)) MiddlewareHandler {
	m.afterFinal = after

	return m
}

func (m MiddlewareHandler) SetProfiler(profiler Profiler) MiddlewareHandler {
	m.profiler = profiler

	return m
}

func (m MiddlewareHandler) Handle(ctx *fasthttp.RequestCtx) {
	if m.profiler != nil {
		m.profiler.Start()
	}

	for _, middleware := range m.middleware {
		ok, e := middleware(ctx)
		if e != nil {
			if m.ErrorHandler != nil {
				m.ErrorHandler.Error(e)
			}
		}
		if !ok {
			return
		}
	}

	if m.final != nil {
		m.final(ctx)
	}

	for _, after := range m.afterFinal {
		ok, e := after(ctx)
		if e != nil {
			if m.ErrorHandler != nil {
				m.ErrorHandler.Error(e)
			}
		}
		if !ok {
			return
		}
	}
	if m.profiler != nil {
		m.profiler.End()
	}
}

func (m MiddlewareHandler) HandleLimited() fasthttp.RequestHandler {
	if m.Limiter != nil {
		return tollbooth_fasthttp.LimitHandler(m.Handle, m.Limiter)
	}

	return m.Handle
}

func HtmlMiddleware(ctx *fasthttp.RequestCtx) (bool, error) {
	ctx.SetContentType("text/html")

	return true, nil
}

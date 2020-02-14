package cbweb

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"log"
	"runtime"
	"strings"
)


type ErrorHandler interface {
	Error(e error)
}

type DefaultErrorHandler struct{}

func (d DefaultErrorHandler) Error(e error) {
	buf := make([]byte, 1000000)
	runtime.Stack(buf, false)
	buf = bytes.Trim(buf, "\x00")
	stack := string(buf)
	stackParts := strings.Split(stack, "\n")
	newStackParts := []string{stackParts[0]}
	newStackParts = append(newStackParts, stackParts[3:]...)
	stack = strings.Join(newStackParts, "\n")
	log.Println("ERROR", e.Error()+"\n"+stack)
}

type MiddlewareHandler struct {
	ErrorHandler ErrorHandler
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
}

func HtmlMiddleware(ctx *fasthttp.RequestCtx) (bool, error) {
	ctx.SetContentType("text/html")

	return true, nil
}
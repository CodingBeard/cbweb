package cbweb

import (
	"errors"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
)

type Module interface {
	SetRoutes(router *router.Router)
	GetGlobalTemplates() map[string][]byte
	SetGlobalTemplates(templates map[string][]byte)
}

type Server struct {
	port               string
	maxRequestBodySize int
	modules            []Module
	errorHandler       ErrorHandler
	globalMiddleware   *MiddlewareHandler
}

type Dependencies struct {
	Port               string
	MaxRequestBodySize int
	ErrorHandler       ErrorHandler
	GlobalMiddleware   *MiddlewareHandler
}

func NewServer(dependencies Dependencies, modules ...Module) *Server {
	if dependencies.ErrorHandler == nil {
		dependencies.ErrorHandler = &DefaultErrorHandler{}
	}
	return &Server{
		port:               dependencies.Port,
		maxRequestBodySize: dependencies.MaxRequestBodySize,
		errorHandler:       dependencies.ErrorHandler,
		globalMiddleware:   dependencies.GlobalMiddleware,
		modules:            modules,
	}
}

func (s *Server) AddModule(module Module) {
	s.modules = append(s.modules, module)
}

func (s *Server) Start() error {
	routes := router.New()
	routes.RedirectTrailingSlash = false
	routes.RedirectFixedPath = false

	globalTemplates := make(map[string][]byte)

	for _, module := range s.modules {
		module.SetRoutes(routes)
		for templateName, templateBytes := range module.GetGlobalTemplates() {
			globalTemplates[templateName] = templateBytes
		}
	}

	for _, module := range s.modules {
		module.SetGlobalTemplates(globalTemplates)
	}

	server := fasthttp.Server{
		MaxRequestBodySize: s.maxRequestBodySize,
		Handler: func(ctx *fasthttp.RequestCtx) {
			defer func() {
				rec := recover()
				if rec != nil {
					if e, ok := rec.(error); ok {
						s.errorHandler.Error(e)
					} else {
						s.errorHandler.Error(errors.New(fmt.Sprint(rec)))
					}
					ctx.Response.SetStatusCode(500)
				}
			}()
			if s.globalMiddleware == nil {
				routes.Handler(ctx)
			} else {
				s.globalMiddleware.SetFinal(routes.Handler).HandleLimited()(ctx)
			}
		},
	}

	e := server.ListenAndServe(s.port)

	return e
}

func (s *Server) RunAndCatch(catch map[os.Signal]func()) {
	var signals []os.Signal
	for toCatch := range catch {
		signals = append(signals, toCatch)
	}
	caught := make(chan os.Signal, 1)
	signal.Notify(caught, signals...)

	go func() {
		e := s.Start()
		if e != nil {
			s.errorHandler.Error(e)
		}
	}()

	sig := <-caught
	if fun, ok := catch[sig]; ok && fun != nil {
		fun()
	}
}

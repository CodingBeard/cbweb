package cbweb

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Module interface {
	SetRoutes(router *router.Router)
	GetGlobalTemplates() map[string][]byte
	SetGlobalTemplates(templates map[string][]byte)
}

type Server struct {
	port             string
	modules          []Module
	errorHandler     ErrorHandler
	globalMiddleware *MiddlewareHandler
}

type Dependencies struct {
	Port             string
	ErrorHandler     ErrorHandler
	GlobalMiddleware *MiddlewareHandler
}

func NewServer(dependencies Dependencies, modules ...Module) *Server {
	if dependencies.ErrorHandler == nil {
		dependencies.ErrorHandler = &DefaultErrorHandler{}
	}
	return &Server{
		port:             dependencies.Port,
		errorHandler:     dependencies.ErrorHandler,
		globalMiddleware: dependencies.GlobalMiddleware,
		modules:          modules,
	}
}

func (s *Server) AddModule(module Module) {
	s.modules = append(s.modules, module)
}

func (s *Server) Start() error {
	routes := router.New()

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

	e := fasthttp.ListenAndServe(s.port, func(ctx *fasthttp.RequestCtx) {
		defer s.errorHandler.Recover()
		if s.globalMiddleware == nil {
			routes.Handler(ctx)
		} else {
			s.globalMiddleware.SetFinal(routes.Handler).HandleLimited()(ctx)
		}
	})

	return e
}

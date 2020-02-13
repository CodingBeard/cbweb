package cbweb

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"html/template"
)

type Module interface {
	SetRoutes(router *router.Router)
	GetGlobalTemplates() map[string][]byte
	SetGlobalTemplates(templates map[string][]byte)
}

type Server struct {
	port string
	modules []Module
}

type ViewIncludeType string

var (
	ViewIncludeType_JsHead           ViewIncludeType = "js-head"
	ViewIncludeType_JsHeadInline     ViewIncludeType = "js-head-inline"
	ViewIncludeType_CssHead          ViewIncludeType = "css-head"
	ViewIncludeType_CssHeadInline    ViewIncludeType = "css-head-inline"
	ViewIncludeType_JsBody           ViewIncludeType = "js-body"
	ViewIncludeType_JsBodyInline     ViewIncludeType = "js-body-inline"
	ViewIncludeType_CssBody          ViewIncludeType = "css-body"
	ViewIncludeType_CssBodyInline    ViewIncludeType = "css-body-inline"
	ViewIncludeType_JsPostBody       ViewIncludeType = "js-postBody"
	ViewIncludeType_JsPostBodyInline ViewIncludeType = "js-postBody-inline"
)

type ViewInclude struct {
	Type      ViewIncludeType
	Src       template.URL
	Html      template.HTML
	Attribute template.HTMLAttr
	Js        template.JS
	Css       template.CSS
}

//todo nav generation
type MasterViewModel interface {
	GetViewIncludes() []ViewInclude
	GetTitle() string
}

// This is here purely for typehinting in go template files
type MasterViewModelTypeHinting struct{}

func (m MasterViewModelTypeHinting) GetViewIncludes() []ViewInclude {
	panic("implement me")
}

func (m MasterViewModelTypeHinting) GetTitle() string {
	panic("implement me")
}

func NewServer(port string, modules ...Module) *Server {
	return &Server{
		port: port,
		modules:modules,
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

	e := fasthttp.ListenAndServe(s.port, routes.Handler)

	return e
}

func (h ViewIncludeType) IsJsHead() bool {
	return h == ViewIncludeType_JsHead
}

func (h ViewIncludeType) IsJsHeadInline() bool {
	return h == ViewIncludeType_JsHeadInline
}

func (h ViewIncludeType) IsCssHead() bool {
	return h == ViewIncludeType_CssHead
}

func (h ViewIncludeType) IsCssHeadInline() bool {
	return h == ViewIncludeType_CssHeadInline
}

func (h ViewIncludeType) IsJsBody() bool {
	return h == ViewIncludeType_JsBody
}

func (h ViewIncludeType) IsJsBodyInline() bool {
	return h == ViewIncludeType_JsBodyInline
}

func (h ViewIncludeType) IsCssBody() bool {
	return h == ViewIncludeType_CssBody
}

func (h ViewIncludeType) IsCssBodyInline() bool {
	return h == ViewIncludeType_CssBodyInline
}

func (h ViewIncludeType) IsJsPostBody() bool {
	return h == ViewIncludeType_JsPostBody
}

func (h ViewIncludeType) IsJsPostBodyInline() bool {
	return h == ViewIncludeType_JsPostBodyInline
}
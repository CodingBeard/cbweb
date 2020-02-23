package cbweb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"html/template"
	"log"
	"runtime"
	"strings"
)

type Module interface {
	SetRoutes(router *router.Router)
	GetGlobalTemplates() map[string][]byte
	SetGlobalTemplates(templates map[string][]byte)
}

type ErrorHandler interface {
	Error(e error)
	Recover()
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

func (d DefaultErrorHandler) Recover() {
	e := recover()

	if e != nil {
		err, ok := e.(error)

		if ok {
			d.Error(err)
		} else {
			d.Error(errors.New(fmt.Sprint(e)))
		}
	}
}

type Server struct {
	port         string
	modules      []Module
	errorHandler ErrorHandler
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

type NavItem struct {
	Src template.URL
	Title string
	Active bool
	SubNavItems []NavItem
	Divider bool
}

type FlashMessage struct {
	Type    string
	Message string
}

type Flash struct {
	Messages map[string][]FlashMessage
}

func (f *Flash) AddMessage(group string, message FlashMessage) {
	if f.Messages == nil {
		f.Messages = make(map[string][]FlashMessage)
	}
	f.Messages[group] = append(f.Messages[group], message)
}

func (f *Flash) GetMessages(group string) []FlashMessage {
	if f.Messages == nil {
		f.Messages = make(map[string][]FlashMessage)
	}

	messages := f.Messages[group]

	delete(f.Messages, group)

	return messages
}

func (f *Flash) HasMessages(group string) bool {
	if f.Messages == nil {
		return false
	}

	return len(f.Messages[group]) > 0
}

type MasterViewModel interface {
	GetMasterViewModel() DefaultMasterViewModel
}

type DefaultMasterViewModel struct {
	ViewIncludes []ViewInclude
	Title string
	PageTitle string
	BodyClasses string
	NavItems []NavItem
	Path template.URL
	Flash *Flash
}

func (m DefaultMasterViewModel) GetViewIncludes() []ViewInclude {
	return m.ViewIncludes
}

func (m DefaultMasterViewModel) GetTitle() string {
	return m.Title
}

func (m DefaultMasterViewModel) GetPageTitle() string {
	return m.PageTitle
}

func (m DefaultMasterViewModel) GetBodyClasses() string {
	return m.BodyClasses
}

func (m DefaultMasterViewModel) GetNavItems() []NavItem {
	return m.NavItems
}

func (m DefaultMasterViewModel) GetPath() template.URL {
	return m.Path
}

// This is here purely for typehinting in go template files
type TypehintingViewModel struct{}

func (t TypehintingViewModel) GetMasterViewModel() DefaultMasterViewModel {
	return DefaultMasterViewModel{}
}

type Dependencies struct {
	Port         string
	ErrorHandler ErrorHandler
}

func NewServer(dependencies Dependencies, modules ...Module) *Server {
	if dependencies.ErrorHandler == nil {
		dependencies.ErrorHandler = &DefaultErrorHandler{}
	}
	return &Server{
		port:         dependencies.Port,
		errorHandler: dependencies.ErrorHandler,
		modules:      modules,
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
		routes.Handler(ctx)
	})

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

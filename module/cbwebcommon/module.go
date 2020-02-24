package cbwebcommon

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/codingbeard/cbweb/templates"
	"github.com/valyala/fasthttp"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Module struct {
	Env              string
	Version          string
	BrandName        string
	FileServer       func(ctx *fasthttp.RequestCtx)
	FourOFourError   func(ctx *fasthttp.RequestCtx)
	FiveHundredError func(ctx *fasthttp.RequestCtx)
	TemplateFuncs    template.FuncMap
	TemplatesBox     FileOpener
	WebAssets        FileOpener
	ErrorHandler     ErrorHandler
	Logger           Logger
	globalTemplates  map[string][]byte
}

type FileOpener interface {
	Open(name string) (http.File, error)
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

type Logger interface {
	InfoF(category string, message string, args ...interface{})
}

type defaultLogger struct{}

func (d defaultLogger) InfoF(category string, message string, args ...interface{}) {
	log.Println(category+":", fmt.Sprintf(message, args...))
}

func (m *Module) SetDefaults() {
	m.FileServer = m.DefaultFileServer
	m.FourOFourError = m.DefaultFourOFourError
	m.FiveHundredError = m.DefaultFiveHundredError
}

func (m *Module) GetFiveHundredError() func(ctx *fasthttp.RequestCtx) {
	return m.FiveHundredError
}

func (m *Module) GetGlobalTemplates() map[string][]byte {
	return map[string][]byte{
		"-global-/cbwebcommon/master.gohtml": getGlobalMasterTemplate(),
		"-global-/cbwebcommon/nav.gohtml": getGlobalNavTemplate(),
		"-global-/cbwebcommon/flash.gohtml": getGlobalFlashTemplate(),
		"-global-/cbwebcommon/inputtext.gohtml": getGlobalInputTextTemplate(),
	}
}

func (m *Module) SetGlobalTemplates(templates map[string][]byte) {
	m.globalTemplates = templates
}

func (m *Module) DefaultFileServer(ctx *fasthttp.RequestCtx) {
	uri := string(ctx.URI().Path())
	if strings.Contains(uri, "?") {
		uri = strings.Split(uri, "?")[0]
	}
	f, e := m.WebAssets.Open(uri)
	if e != nil {
		if strings.Contains(e.Error(), "file does not exist") || strings.Contains(e.Error(), "no such file or directory") {
			m.FourOFourError(ctx)
			return
		}
		m.ErrorHandler.Error(e)
		m.FiveHundredError(ctx)
		return
	}
	stat, e := f.Stat()
	if e != nil {
		m.ErrorHandler.Error(e)
		m.FiveHundredError(ctx)
		return
	}
	if !ctx.IfModifiedSince(stat.ModTime()) {
		ctx.NotModified()
		return
	}
	ctx.Response.Header.SetLastModified(stat.ModTime())
	ctx.Response.Header.Set("Cache-Control", "max-age=31536000")
	ctx.Response.Header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(stat.Name())))
	ctx.Response.Header.SetContentLength(int(stat.Size()))
	ctx.SetBodyStream(f, int(stat.Size()))
}

func (m *Module) GenerateTemplate(fileNames []string) (*templates.InheritanceMultiTemplate, error) {
	mergedTemplateFuncs := m.getDefaultTemplateFuncs()
	for key, templateFunc := range m.TemplateFuncs {
		mergedTemplateFuncs[key] = templateFunc
	}

	t := templates.NewInheritanceMultiTemplate(mergedTemplateFuncs)

	for templateName, templateBytes := range m.globalTemplates {
		e := t.AddTemplate(templateName, templateBytes)
		if e != nil {
			m.ErrorHandler.Error(e)
		}
	}

	if len(fileNames) == 0 {
		return nil, errors.New("no fileNames provided")
	} else {
		for _, fileName := range fileNames {
			templateFile, e := m.TemplatesBox.Open(fileName)
			if e != nil {
				m.ErrorHandler.Error(e)
				return nil, e
			}

			templateBytes, e := ioutil.ReadAll(templateFile)
			if e != nil {
				m.ErrorHandler.Error(e)
				return nil, e
			}

			e = t.AddTemplate(fileName, templateBytes)
			if e != nil {
				m.ErrorHandler.Error(e)
				return nil, e
			}
		}

	}

	return t, nil
}

func (m *Module) getDefaultTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"getCdnUrlString":      m.getDefaultCdnUrl,
		"getCdnUrlTemplateURL": m.getDefaultCdnUrlTemplateUrl,
		"getVersionString":     m.getDefaultVersionString,
		"getBrandName":         m.getDefaultBrandName,
	}
}

func (m *Module) getDefaultCdnUrlTemplateUrl(nonCdnUrl template.URL) string {
	return m.getDefaultCdnUrl(string(nonCdnUrl))
}

func (m *Module) getDefaultCdnUrl(nonCdnUrl string) string {
	if m.Env == "dev" {
		return nonCdnUrl + "?" + strconv.Itoa(int(time.Now().Unix()))
	} else {
		//todo cdn
		return nonCdnUrl + "?" + m.Version
	}
}

func (m *Module) getDefaultVersionString() string {
	return m.Version
}

func (m *Module) getDefaultBrandName() string {
	return m.BrandName
}

func (m *Module) DefaultFiveHundredError(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(500)
	_, _ = fmt.Fprint(ctx, `Error: 500`)
}

func (m *Module) DefaultFourOFourError(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(404)
	_, _ = fmt.Fprint(ctx, `Error: 404 Not Found`)
}

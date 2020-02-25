package templates

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/codingbeard/cbweb"
	"html/template"
	"io"
	"regexp"
	"time"
)

var extendsRegex *regexp.Regexp

func init() {
	var e error
	extendsRegex, e = regexp.Compile(`\{\{ *?extends +?"(.+?)" *?\}\}`)
	if e != nil {
		panic(e)
	}
}

type templatefile struct {
	layout  string
	content []byte
}

type InheritanceMultiTemplate struct {
	templates       map[string]templatefile
	funcs           template.FuncMap
	cache           bool
	cachedTemplates cbweb.CacheProvider
}

type Dependencies struct {
	Funcs         template.FuncMap
	Cache         bool
	CacheProvider cbweb.CacheProvider
}

func NewInheritanceMultiTemplate(dependencies Dependencies) *InheritanceMultiTemplate {
	return &InheritanceMultiTemplate{
		templates:       make(map[string]templatefile),
		funcs:           dependencies.Funcs,
		cache:           dependencies.Cache,
		cachedTemplates: dependencies.CacheProvider,
	}
}

func (m *InheritanceMultiTemplate) AddTemplate(name string, content []byte) error {
	if m.cache {
		if _, ok := m.templates[name]; ok {
			return nil
		}
	}

	templ := templatefile{
		content: content,
	}

	r := bytes.NewReader(templ.content)
	pos := 0
	var line []byte
	for {
		ch, l, err := r.ReadRune()
		pos += l

		// read until first line or EOF
		if ch == '\n' || err == io.EOF {
			line = templ.content[0:pos]
			break
		}
	}

	if len(line) < 10 {
		m.templates[name] = templ
		return nil
	}

	// if we have a match, strip first line of content
	if m := extendsRegex.FindSubmatch(line); m != nil {
		templ.layout = string(m[1])
		templ.content = templ.content[len(line):]
	}

	m.templates[name] = templ

	return nil
}

func (m *InheritanceMultiTemplate) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	var t *template.Template
	var ok bool
	cacheKey := "executeTemplate:" + name
	if m.cachedTemplates != nil {
		var cache interface{}
		cache, ok = m.cachedTemplates.Get(cacheKey)
		if ok {
			t = cache.(*template.Template)
		}
	}
	if !m.cache || !ok {
		templ, ok := m.templates[name]
		if !ok {
			return errors.New(fmt.Sprintf("template (%s) does not exist", name))
		}
		t = template.New(name)
		if templ.layout != "" {
			layout, ok := m.templates[templ.layout]
			if !ok {
				return errors.New(fmt.Sprintf("template layout (%s) does not exist", templ.layout))
			}

			templ.content = append(templ.content, layout.content...)
		}
		t.Funcs(m.funcs)
		var e error
		t, e = t.Parse(string(templ.content))
		if e != nil {
			return e
		}

		for templName, templ := range m.templates {
			if templName == name {
				continue
			}
			if templ.layout != "" {
				layout, ok := m.templates[templ.layout]
				if !ok {
					return errors.New(fmt.Sprintf("template layout (%s) does not exist", templ.layout))
				}

				templ.content = append(templ.content, layout.content...)
			}
			var e error
			t, e = t.Parse(string(templ.content))
			if e != nil {
				return e
			}
		}

		if m.cachedTemplates != nil {
			m.cachedTemplates.Set(cacheKey, t, time.Hour*24)
		}
	}

	return t.Execute(wr, data)
}

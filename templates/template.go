package templates

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"regexp"
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
	templates map[string]templatefile
	funcs     template.FuncMap
}

func NewInheritanceMultiTemplate(funcs template.FuncMap) *InheritanceMultiTemplate {
	return &InheritanceMultiTemplate{
		templates: make(map[string]templatefile),
		funcs:     funcs,
	}
}

func (m *InheritanceMultiTemplate) AddTemplate(name string, content []byte) error {
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
	templ, ok := m.templates[name]
	if !ok {
		return errors.New(fmt.Sprintf("template (%s) does not exist", name))
	}
	t := template.New(name)
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

	return t.Execute(wr, data)
}
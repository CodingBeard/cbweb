package cbform

import "strings"

type Container struct {
	fields map[string]*Field
}

type Field struct {
	Name                string
	Type                string
	Placeholder         string
	Label               string
	JsValidationError   string
	JsValidationSuccess string
	Error               error
	Help                string
	Value               interface{}
}

func New(fields ...Field) *Container {
	mapFields := make(map[string]*Field, len(fields))

	for key := range fields {
		mapFields[fields[key].Name] = &fields[key]
	}

	return &Container{fields: mapFields}
}

func (c *Container) GetField(fieldName string) *Field {
	field, ok := c.fields[fieldName]
	if ok {
		return field
	}

	return &Field{}
}

func (f *Field) GetName() string {
	return f.Name
}

func (f *Field) GetType() string {
	return f.Type
}

func (f *Field) GetPlaceholder() string {
	return f.Placeholder
}

func (f *Field) GetLabel() string {
	return f.Label
}

func (f *Field) GetJsValidationError() string {
	return f.JsValidationError
}

func (f *Field) GetJsValidationSuccess() string {
	return f.JsValidationSuccess
}

func (f *Field) GetError() string {
	if f.Error != nil {
		return strings.Title(f.Error.Error())
	}

	return ""
}

func (f *Field) HasError() bool {
	return f.Error != nil
}

func (f *Field) GetHelp() string {
	return f.Help
}

func (f *Field) GetValue() interface{} {
	return f.Value
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
)

//** Render context **

type RenderContext struct {
	templates       map[string]string
	predefinedTypes map[string]string
	abbrs           []string
}

func (c *RenderContext) array() (out []string) {
	out = make([]string, 0)
	for _, v := range c.templates {
		out = append(out, v)
	}
	return
}

func (c *RenderContext) find(key string) (name, path string, err error) {

	path, ok := c.templates[key]
	if !ok {
		err = errors.New("invalid type")
		return
	}

	i := strings.LastIndex(path, "/")
	if i < 0 {
		name = path
	} else {
		name = path[i+1:]
	}

	return
}

var renderContext = &RenderContext{
	templates: map[string]string{
		"client":             "templates/client/client.tmpl",
		"handlers":           "templates/server/handlers.tmpl",
		"client_signature":   "templates/client/signature.tmpl",
		"handlers_signature": "templates/server/signature.tmpl",
		"object":             "templates/base/struct.tmpl",
		"array":              "templates/base/array.tmpl",
	},
	predefinedTypes: map[string]string{
		"string":  "string",
		"integer": "int64",
		"number":  "float64",
		"boolean": "bool",
	},
	abbrs: []string{"id", "href", "url"},
}

//** Render Schema **

func (s *Schema) RenderType() string {
	simpleType, ok := renderContext.predefinedTypes[s.Type]
	if ok {
		return simpleType
	}
	switch s.Type {
	case "object":
		return convertToGoName(s.Name, false)
	case "array":
		if s.Name == "" {
			if s.Items.Name == "" {
				return "[]" + s.Items.Type
			}

			return "[]*" + convertToGoName(s.Items.Name, false)
		}

		return "*" + s.Name
	default:
		panic("invalid type")
	}
}

// RenderDefinition renders Schema as struct definition.
func (s *Schema) RenderDefinition() string {
	simpleType, ok := renderContext.predefinedTypes[s.Type]
	if ok {
		return simpleType
	}

	name, filepath, err := renderContext.find(s.Type)
	if err != nil {
		panic(err)
	}

	return s.render(name, filepath)
}

func (s *Schema) render(name, filename string) string {

	buf := bytes.NewBuffer([]byte{})

	tmpl, err := template.New(name).Funcs(getFuncMap()).ParseFiles(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = tmpl.Execute(buf, s)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return buf.String()
}

//** Operation **
func (o *Operation) GetBodyName(method string) *string {

	switch method {
	case "POST", "PUT", "PATCH":
		for _, el := range o.Parameters {
			if el.In == "body" {
				return &el.Name
			}
		}
	}
	return nil
}

//** Render Parameter **

func (p *Parameter) RenderAsValue() string {
	if !p.Required {
		return "*" + p.Name
	}
	return p.Name
}

func (p *Parameter) RenderType() string {
	if !p.Required {
		return "*" + p.Schema.RenderType()
	}
	return p.Schema.RenderType()
}

// RenderAsString converts variable to string if needed.
func (p *Parameter) RenderAsString() (out string) {
	if p.Schema.Type != "string" {
		return fmt.Sprintf("strconv.FormatInt(%s, 10)", p.RenderAsValue())
	}
	switch p.Schema.Type {
	case "integer":
		out = fmt.Sprintf("strconv.FormatInt(%s, 10)", p.RenderAsValue())
	case "number":
		out = fmt.Sprintf("strconv.FormatFloat(%s, 'f', -1, 64)", p.RenderAsValue())
	default:
		out = p.RenderAsValue()
	}
	return
}

func (p *Parameter) RenderValueTo(value string) (out string) {
	switch p.Schema.Type {
	case "integer":
		out = fmt.Sprintf("%s, err = strconv.ParseInt(%s, 10, 64)", p.RenderAsValue(), value)
	case "number":
		out = fmt.Sprintf("%s, err =  strconv.ParseFloat(%s, 64)", p.RenderAsValue(), value)
	default:
		out = fmt.Sprintf("%s = %s", p.RenderAsValue(), value)
	}
	return
}

func (p *Parameter) IsReturnsParsingError() bool {
	if p.Schema.Type == "integer" || p.Schema.Type == "number" {
		return true
	}
	return false
}

//** Render Swagger **

func (s *Swagger) String() string {
	b, _ := json.MarshalIndent(s, "", "  ")

	return string(b)
}

func render(s *Swagger, tmplName string) {
	tmpl, err := template.New(tmplName).Funcs(getFuncMap()).ParseFiles(renderContext.array()...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = tmpl.Execute(os.Stdout, s)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

//** Utils **

func convertToGoName(name string, varname bool) string {
	split := func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	}

	names := []string{}

	for i, name := range strings.FieldsFunc(name, split) {
		if contains(renderContext.abbrs, name) {
			names = append(names, strings.ToUpper(name))
		} else if i == 0 && varname {
			names = append(names, name)
		} else {
			names = append(names, capitalize(name))
		}
	}

	return strings.Join(names, "")
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"goName": func(name string, varname bool) string {
			return convertToGoName(name, varname)
		},
		"contains": func(s []string, item string) bool {
			return contains(s, item)
		},
	}
}

func contains(s []string, item string) bool {
	for _, v := range s {
		if v == item {
			return true
		}
	}
	return false
}

func capitalize(s string) string {
	return strings.ToUpper(string(s[0])) + string(s[1:])
}

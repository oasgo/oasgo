package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"errors"

	"github.com/gobuffalo/packr"
)

//** Render context **

type RenderContext struct {
	Box             *packr.Box
	templates       map[string]string
	predefinedTypes map[string]string
	abbrs           []string
}

func (c *RenderContext) find(key string) (name, tpl string, err error) {

	name = key
	tpl, ok := c.templates[key]
	if !ok {
		err = errors.New("invalid type")
		return
	}
	return
}

func NewRenderContext() *RenderContext {
	box := packr.NewBox("./templates")
	return &RenderContext{
		Box: &box,
		templates: map[string]string{
			"client":             box.String("client/client.tmpl"),
			"handlers":           box.String("server/handlers.tmpl"),
			"client_signature":   box.String("client/signature.tmpl"),
			"handlers_signature": box.String("server/signature.tmpl"),
			"object":             box.String("base/struct.tmpl"),
			"array":              box.String("base/array.tmpl"),
		},
		predefinedTypes: map[string]string{
			"string":  "string",
			"integer": "int64",
			"number":  "float64",
			"boolean": "bool",
		},
		abbrs: []string{"id", "href", "url"},
	}
}

var renderContext = NewRenderContext()

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

	name, tpl, err := renderContext.find(s.Type)
	if err != nil {
		panic(err)
	}

	return s.render(name, tpl)
}

func (s *Schema) render(name, tpl string) string {

	buf := bytes.NewBuffer([]byte{})

	tmpl, err := template.New(name).Funcs(getFuncMap()).Parse(tpl)
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
				return &(el.Name)
			}
		}
	}
	return nil
}

//** Render Parameter **

func (p *Parameter) RenderAsValue() string {
	if !p.Required {
		return "*" + p.ToCamelCase()
	}
	return p.ToCamelCase()
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

func (p *Parameter) ToCamelCase() (out string) {

	bs := []byte(p.Name)
	bs = regexp.MustCompile(`([a-zA-Z])(\d+)([a-zA-Z]?)`).ReplaceAll(bs, []byte(`$1 $2 $3`))
	in := strings.Trim(string(bs), " ")

	isNext := false
	for _, v := range in {
		if v >= 'A' && v <= 'Z' {
			out += string(v)
		}
		if v >= '0' && v <= '9' {
			out += string(v)
		}
		if v >= 'a' && v <= 'z' {
			if isNext {
				out += strings.ToUpper(string(v))
			} else {
				out += string(v)
			}
		}
		if v == '_' || v == ' ' || v == '-' {
			isNext = true
		} else {
			isNext = false
		}
	}
	return
}

//** Render Swagger **

func (s *Swagger) String() string {
	b, _ := json.MarshalIndent(s, "", "  ")

	return string(b)
}

func render(s *Swagger, tmplName string) {

	_, tpl, err := renderContext.find(tmplName)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New(tmplName).Funcs(getFuncMap()).Parse(tpl)
	if err != nil {
		os.Stderr.WriteString("Parse tmpl error: " + err.Error())
		os.Exit(1)
	}

	for k, v := range renderContext.templates {
		if k != tmplName {
			associated, err := template.New(k).Funcs(getFuncMap()).Parse(v)
			if err != nil {
				os.Stderr.WriteString("Parse associated tmpl error: " + err.Error())
				os.Exit(3)
			}
			tmpl, err = tmpl.AddParseTree(k, associated.Tree)
			if err != nil {
				os.Stderr.WriteString("AddParseTree tmpl error: " + err.Error())
				os.Exit(4)
			}
		}
	}

	err = tmpl.ExecuteTemplate(os.Stdout, tmplName, s)
	if err != nil {
		os.Stderr.WriteString("Execute tmpl error: " + err.Error())
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

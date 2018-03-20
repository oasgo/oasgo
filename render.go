package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

var abbrs = []string{"id", "href", "url"}

func (s *Schema) renderAsObject() string {
	buf := bytes.NewBuffer([]byte{})
	tmpl, err := template.New("struct.tmpl").Funcs(getFuncMap()).ParseFiles("templates/struct.tmpl")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = tmpl.Execute(os.Stdout, s)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return buf.String()
}

func (s *Schema) renderAsArray() string {
	buf := bytes.NewBuffer([]byte{})
	tmpl, err := template.New("array.tmpl").Funcs(getFuncMap()).ParseFiles("templates/array.tmpl")
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

func (s *Schema) RenderType() string {
	simpleType, ok := predefinedTypes[s.Type]
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
func (p *Parameter) RenderAsString() string {
	if p.Schema.Type != "string" {
		return fmt.Sprintf("strconv.Itoa(%s)", p.RenderAsValue())
	}

	return p.RenderAsValue()
}

// RenderDefinition renders Schema as struct definition.
func (s *Schema) RenderDefinition() string {
	simpleType, ok := predefinedTypes[s.Type]
	if ok {
		return simpleType
	}
	switch s.Type {
	case "object":
		return s.renderAsObject()
	case "array":
		return s.renderAsArray()
	default:
		panic("invalid type")
	}
}

func (s *Swagger) String() string {
	b, _ := json.MarshalIndent(s, "", "  ")

	return string(b)
}

func render(s *Swagger) {
	tmpl, err := template.New("package.tmpl").Funcs(getFuncMap()).ParseFiles(templates...)
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

func convertToGoName(name string, varname bool) string {
	split := func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	}

	names := []string{}

	for i, name := range strings.FieldsFunc(name, split) {
		if contains(abbrs, name) {
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

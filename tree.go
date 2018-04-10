package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

const (
	baseTemplate = `
/*
* This file autogenerated;
*
* DO NOT EDIT
* Generated by https://github.com/oasgo/oasgo
*/

package {{.PackageName}}
type (
	{{ range $r := .References }}
		{{ $r.Reference.RenderDefinition }}
	{{ end }}
)

{{ range $f := .Functions -}}
func {{ $f.RenderSignature }} {
		{{ $f.RenderBody }}
	}
{{ end }}
`
	structTemplate = `{{$.Name}} struct {
	{{ range $p :=  $.Properties }}
	{{- $p.Name }} {{ $p.Reference.RenderName }} {{($.RenderTags $p)}}
	{{ end }}
}`
	arrayTemplate     = `{{$.Name}} []{{$.ItemsType.Reference.RenderName}}`
	signatureTemplate = `` +
		`{{$.Name}}(r *http.Request)` +
		`({{range $i, $a := $.Output}}{{$a.Property.Name}} ` +
		`{{$a.Property.Reference.RenderName}}{{- if lt (inc $i) (len $.Output) -}}, {{- end -}}` +
		`{{end}})`
	funcBodyTemplate = `
{{ if gt (len $.Output) 0 }}
var value string
{{ range $i, $p := $.Output }}
{{$p.RenderExtraction}}
{{ end }}
{{ end }}
return
`
	paramTemplate = `
r.URL.Query().Get("{{- $.Property.SourceName}}")
{{- if $.Required }}
	if value == "" {
		err = &MissingParameterError{field:  "{{- $.Property.SourceName}}"}
		return
	}
{{- end }}
{{(($.Property.Reference.RenderExtraction $.Property.Name $.Property.SourceName))}}
`
	extractIntTemplate = `
{{$.Name}}, err = strconv.ParseInt(value, 10, 64)
if err != nil {
	err = &InvalidParameterTypeError{
		field:"{{$.Field}}",
		original: err,
	}
	return
}
`
	extractFloatTemplate = `
{{$.Name}}, err = strconv.ParseFloat(value, 64)
if err != nil {
	err = &InvalidParameterTypeError{
		field:"{{$.Field}}",
		original: err,
	}
	return
}
`
	extractSliceTemplate  = ``
	extractStructTemplate = ``
)

type Reference interface {
	RenderDefinition() string
	RenderLiteral() string
	RenderName() string
	RenderExtraction(varName, oName string) string
}

type Context struct {
	PackageName string
	References  map[string]property
	Functions   []Function
}
type Function struct {
	Name   string
	Output []Param
}

type Param struct {
	In       string
	Required bool
	Property property
}

type Struct struct {
	Properties []property
	Name       string
}

type Slice struct {
	ItemsType property
	Name      string
}

type String struct{}
type Integer struct{}
type Number struct{}

type property struct {
	Name       string
	SourceName string
	Reference  Reference
	Required   bool
}

func renderTemplate(tname, t string, i interface{}) string {
	buf := bytes.NewBuffer([]byte{})

	tmpl, err := template.New(tname).Funcs(template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}).Parse(t)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = tmpl.Execute(buf, i)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return buf.String()

}

func (f *Function) RenderSignature() string {
	return renderTemplate("signature", signatureTemplate, f)
}

func newParam(in string, required bool, p property) Param {
	p.Name = ToCamelCase(false, p.Name)
	return Param{in, required, p}
}

func (f *Function) RenderBody() string {
	return renderTemplate("funcBody", funcBodyTemplate, f)
}

func (s *String) RenderLiteral() string    { return "string" }
func (s *String) RenderName() string       { return "string" }
func (s *String) RenderDefinition() string { return "" }
func (s *String) RenderExtraction(vn, on string) string {
	return fmt.Sprintf("%s = value", vn)
}

func (i *Integer) RenderLiteral() string    { return "int64" }
func (i *Integer) RenderName() string       { return "int64" }
func (i *Integer) RenderDefinition() string { return "" }
func (i *Integer) RenderExtraction(vn, on string) string {
	return renderTemplate(
		"int", extractIntTemplate,
		struct {
			Name  string
			Field string
		}{vn, on})
}

func (n *Number) RenderLiteral() string    { return "float64" }
func (n *Number) RenderName() string       { return "float64" }
func (n *Number) RenderDefinition() string { return "" }
func (n *Number) RenderExtraction(vn, on string) string {
	return renderTemplate(
		"float", extractFloatTemplate,
		struct {
			Name  string
			Field string
		}{vn, on})
}

func (s *Slice) RenderLiteral() string    { return s.Name }
func (s *Slice) RenderName() string       { return "[]" + s.ItemsType.Reference.RenderName() }
func (s *Slice) RenderDefinition() string { return renderTemplate("slice", arrayTemplate, s) }
func (s *Slice) RenderExtraction(vn, on string) string {
	return renderTemplate("sliceExtract", extractSliceTemplate, s)
}

func (s *Struct) RenderLiteral() string    { return s.Name }
func (s *Struct) RenderName() string       { return s.Name }
func (s *Struct) RenderDefinition() string { return renderTemplate("struct", structTemplate, s) }
func (s *Struct) RenderExtraction(vn, on string) string {
	return renderTemplate("structExtract", extractStructTemplate, s)
}
func (s *Struct) RenderTags(p property) string {
	tags := []string{fmt.Sprintf("json:\"%s\"", p.SourceName)}
	if p.Required {
		tags = append(tags, "valid:\"required\"")
	}
	return fmt.Sprintf("`%s`", strings.Join(tags, " "))
}

func (p *Param) RenderExtraction() string {
	return renderTemplate("paramExtract", paramTemplate, p)
}

func (ctx *Context) setProperty(schema *Schema, name, pname string) property {
	refName := ToCamelCase(true, pname, name)
	if p, ok := ctx.References[refName]; ok {
		return p
	}

	p := property{Name: ToCamelCase(true, name), SourceName: name}
	switch schema.Type {
	case "object":
		ps := &Struct{Name: refName, Properties: []property{}}
		for n, s := range schema.Properties {
			p := ctx.setProperty(s, n, refName)
			for _, a := range schema.Required {
				if n == a {
					p.Required = true
				}
			}
			ps.Properties = append(ps.Properties, p)
		}
		p.Reference = ps
		ctx.References[refName] = p
	case "string":
		p.Reference = &String{}
	case "integer":
		p.Reference = &Integer{}
	case "array":
		p.Reference = &Slice{Name: refName, ItemsType: ctx.setProperty(schema.Items, "", refName)}
	case "number":
		p.Reference = &Number{}
	default:
		log.Fatalf("unsupported type %s", schema.Type)
	}
	return p
}

func (ctx *Context) getParams(ps []*Parameter, opID string) []Param {
	inputs := []Param{}
	for _, p := range ps {
		inputs = append(inputs, newParam(p.In, p.Required, ctx.setProperty(p.Schema, p.ExternalName, opID)))
	}
	return inputs
}

func renderTree(s *Swagger) {

	tmpl, err := template.New("handlers").Parse(baseTemplate)
	if err != nil {
		os.Stderr.WriteString("Parse tmpl error: " + err.Error())
		os.Exit(1)
	}

	c := Context{
		PackageName: "mypackage",
		References:  make(map[string]property),
		Functions:   []Function{},
	}
	for n, schema := range s.Components.Schemas {
		c.setProperty(schema, n, "")
	}

	for _, m := range s.Paths {
		if m.GET != nil {
			c.Functions = append(c.Functions, Function{Name: ToCamelCase(true, m.GET.OperationID), Output: c.getParams(m.GET.Parameters, m.GET.OperationID)})
		}
		if m.POST != nil {
			c.Functions = append(c.Functions, Function{Name: ToCamelCase(true, m.POST.OperationID), Output: c.getParams(m.POST.Parameters, m.POST.OperationID)})
		}
		if m.PUT != nil {
			c.Functions = append(c.Functions, Function{Name: ToCamelCase(true, m.PUT.OperationID), Output: c.getParams(m.PUT.Parameters, m.PUT.OperationID)})
		}
		if m.DELETE != nil {
			c.Functions = append(c.Functions, Function{Name: ToCamelCase(true, m.DELETE.OperationID), Output: c.getParams(m.DELETE.Parameters, m.DELETE.OperationID)})
		}
	}

	err = tmpl.Execute(os.Stdout, c)
	if err != nil {
		os.Stderr.WriteString("Execute tmpl error: " + err.Error())
		os.Exit(2)
	}
}
func ToCamelCase(upper bool, names ...string) (out string) {
	bs := []byte(strings.Join(names, "_"))
	in := strings.Trim(string(bs), " ")

	isNext := false
	for i, v := range in {
		if (v >= 'A' && v <= 'Z') || (v >= 'a' && v <= 'z') {
			if i == 0 {
				if !upper {
					out += strings.ToLower(string(v))
				} else {
					out += strings.ToUpper(string(v))
				}
			} else {
				if isNext {
					out += strings.ToUpper(string(v))
				} else {
					out += string(v)
				}
			}
		}
		if v >= '0' && v <= '9' {
			out += string(v)
		}
		if v == '_' || v == ' ' || v == '-' {
			isNext = true
		} else {
			isNext = false
		}
	}
	return
}
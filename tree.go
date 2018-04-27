package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

const (
	structTemplate = `
	{{- if $.IsAbbreviate}} 
		//{{ $.P.Desc }}
		{{ $.P.AbbrName }} 
	{{- else -}} 
		{{ $.P.Name -}} 
	{{ end }} struct {
		{{ range $p :=  $.P.SortedProperties }}
			{{- $p.Name }} {{ $p.Reference.RenderName $.IsAbbreviate}} {{($.P.RenderTags $p)}}
		{{ end }}
	}
	`
	arrayTemplate     = `{{$.Name}} []{{$.ItemsType.Reference.RenderName false}}`
	dictTemplate      = `{{$.Name}} map[string]{{$.ItemsType.Reference.RenderName false}}`
	signatureTemplate = `{{$.Name}} ( res interface{},
	{{- range $i, $p := $.Input }}
		{{- if eq $p.In "body"}} body {{ else }} {{ $p.Property.Name }} {{ end -}}	
		{{ $p.Property.Reference.RenderName false}} {{- if lt (inc $i) (len $.Input) -}}, {{- end -}}
	{{- end -}}
	)(*http.Response, error)`
	funcBodyTemplate = `
	{{ if $.HasPathParam }}
		c.URL.Path = strings.NewReplacer(
			{{ $.RenderPathParams }}
		).Replace("{{- $.Path -}}")
		{{$.RenderQueryParams}}
	{{ else }}
		c.URL.Path = "{{- $.Path -}}"
	{{ end }}
	{{$.RenderRequestBody}}
`
	pathParamsTemplate = `
    {{- range $p := $.GetPathParams}}
		{{- $p.RenderPathParam}}
    {{- end }}

`
	queryParamsTemplate = `
	q := c.URL.Query()
    {{ range $p := $.GetQueryParams}}
		{{- $p.RenderQueryParam}}
    {{- end }}
    c.URL.RawQuery = q.Encode()
`
	requestBodyTemplate = `
	{{$body := $.GetBody}}
	{{if $body}}
		bs, err := json.Marshal(body)
      	if err != nil {
        	return nil, err
      	}
      	request, err := http.NewRequest("{{$.OperationType.String}}", c.URL.String(), bytes.NewBuffer(bs))
  	{{- else}}
      	request, err := http.NewRequest("{{$.OperationType.String}}", c.URL.String(), nil)
	{{end}}
	if err != nil {
		return nil, err
	}

	return c.sendRequest(res, request)`

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
	setPathParamTemplate = `
	{{- if not $.Required }}
		if {{- $.Property.Name}} != "" {
	{{- end -}}
	"{{"{"}}{{- $.Property.SourceName}}{{"}"}}", {{$.Property.Name}},
	{{- if not $.Required }}
		}
	{{- end  }}
`
	setQueryParamTemplate = `
	{{- if not $.Required }}
		if {{- $.Property.Name}} != "" {
	{{- end -}}
		q.Set("{{- $.Property.SourceName}}", {{- $.Property.Name}})
	{{- if not $.Required }}
		}
	{{- end  }}
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
	extractBoolTemplate = `
	{{$.Name}}, err = strconv.ParseBool(value)
	if err != nil {
		err = &InvalidParameterTypeError{
			field:"{{$.Field}}",
			original: err,
		}
		return
	}
`
	extractSliceTemplate  = ``
	extractDictTemplate   = ``
	extractStructTemplate = ``

	structValidateTemplate = `
		{{- if eq $.Name "" -}}
			return govalidator.ValidateStruct(r)
		{{- else -}}
			return r.{{$.Name}}.Validate()
		{{- end -}}
	`
)

const (
	GET OperationType = iota
	POST
	PUT
	PATCH
	DELETE
)

var operationTypeValues = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

type OperationType int

type Reference interface {
	RenderDefinition(isAbbreviate bool) string
	RenderLiteral() string
	RenderName(isAbbreviate bool) string
	RenderExtraction(varName, oName string) string
	RenderFormat() string
}

type Context struct {
	PackageName  string
	Info         Info
	IsAbbreviate bool
	References   map[string]property
	Functions    []Function
}
type Function struct {
	Name          string
	Path          string
	OperationType OperationType
	Input         []Param
	Output        []Param
}

type Param struct {
	In       string
	Required bool
	Property property
}

type Struct struct {
	Properties []property
	Name       string
	AbbrName   string
	Desc       string
}

type Slice struct {
	ItemsType property
	Name      string
}

type Dictionary struct {
	ItemsType property
	Name      string
}

type String struct {
	Values  []string
	Default string
	Format  string
}

type Integer struct{}
type Number struct{}
type Bool struct{}

type property struct {
	Name       string
	SourceName string
	Reference  Reference
	Required   bool
	Enum       []string
}

type functions []Function
type propertiesByLiteral []property
type propertiesByName []property

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

func (c Context) SortedFunctions() []Function {
	sort.Sort(functions(c.Functions))
	return c.Functions
}

func (c Context) SortedReferences() []property {
	arr := []property{}
	for _, v := range c.References {
		arr = append(arr, v)
	}
	sort.Sort(propertiesByLiteral(arr))
	return arr
}

func (c Struct) SortedProperties() []property {
	sort.Sort(propertiesByName(c.Properties))
	return c.Properties
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

func (f *Function) RenderPathParams() string {
	return renderTemplate("pathParams", pathParamsTemplate, f)
}

func (f *Function) RenderQueryParams() string {
	return renderTemplate("queryParams", queryParamsTemplate, f)
}

func (f *Function) RenderRequestBody() string {
	return renderTemplate("requestBody", requestBodyTemplate, f)
}

func (f *Function) GetBody() *Param {
	for _, el := range f.Input {
		if el.In == "body" {
			return &el
		}
	}
	return nil
}

func (f *Function) HasPathParam() bool {
	for _, el := range f.Input {
		if el.In == "path" {
			return true
		}
	}
	return false
}

func (f *Function) GetPathParams() (params []Param) {
	return f.getParamsByIn("path")
}

func (f *Function) GetQueryParams() (params []Param) {
	return f.getParamsByIn("query")
}

func (f *Function) getParamsByIn(in string) (params []Param) {
	params = make([]Param, 0)
	for _, el := range f.Input {
		if el.In == in {
			params = append(params, el)
		}
	}
	return
}

func (p *Param) RenderPathParam() string {
	return renderTemplate("pParam", setPathParamTemplate, p)
}

func (p *Param) RenderQueryParam() string {
	return renderTemplate("qParam", setQueryParamTemplate, p)
}

func (s *String) RenderLiteral() string                     { return "string" }
func (s *String) RenderName(isAbbreviate bool) string       { return "string" }
func (s *String) RenderDefinition(isAbbreviate bool) string { return "" }
func (s *String) RenderExtraction(vn, on string) string {
	return fmt.Sprintf("%s = value", vn)
}
func (s *String) RenderFormat() string   { return s.Format }
func (s *String) RenderValues() []string { return s.Values }
func (s *String) RenderDefault() string  { return s.Default }

func (i *Integer) RenderLiteral() string                     { return "int64" }
func (i *Integer) RenderName(isAbbreviate bool) string       { return "int64" }
func (i *Integer) RenderDefinition(isAbbreviate bool) string { return "" }
func (i *Integer) RenderExtraction(vn, on string) string {
	return renderTemplate(
		"int", extractIntTemplate,
		struct {
			Name  string
			Field string
		}{vn, on})
}
func (i *Integer) RenderFormat() string { return "" }

func (n *Number) RenderLiteral() string                     { return "float64" }
func (n *Number) RenderName(isAbbreviate bool) string       { return "float64" }
func (n *Number) RenderDefinition(isAbbreviate bool) string { return "" }
func (n *Number) RenderExtraction(vn, on string) string {
	return renderTemplate(
		"float", extractFloatTemplate,
		struct {
			Name  string
			Field string
		}{vn, on})
}
func (n *Number) RenderFormat() string { return "" }

func (b *Bool) RenderLiteral() string                     { return "bool" }
func (b *Bool) RenderName(isAbbreviate bool) string       { return "bool" }
func (b *Bool) RenderDefinition(isAbbreviate bool) string { return "" }
func (b *Bool) RenderExtraction(vn, on string) string {
	return renderTemplate(
		"bool", extractBoolTemplate,
		struct {
			Name  string
			Field string
		}{vn, on})
}
func (b *Bool) RenderFormat() string { return "" }

func (s *Slice) RenderLiteral() string { return s.Name }
func (s *Slice) RenderName(isAbbreviate bool) string {
	return "[]" + s.ItemsType.Reference.RenderName(isAbbreviate)
}
func (s *Slice) RenderDefinition(isAbbreviate bool) string {
	return renderTemplate("slice", arrayTemplate, s)
}
func (s *Slice) RenderExtraction(vn, on string) string {
	return renderTemplate("sliceExtract", extractSliceTemplate, s)
}
func (s *Slice) RenderFormat() string { return "" }

func (s *Dictionary) RenderLiteral() string { return s.Name }
func (s *Dictionary) RenderName(isAbbreviate bool) string {
	return "map[string]" + s.ItemsType.Reference.RenderName(isAbbreviate)
}
func (s *Dictionary) RenderDefinition(isAbbreviate bool) string {
	return renderTemplate("dict", dictTemplate, s)
}
func (s *Dictionary) RenderExtraction(vn, on string) string {
	return renderTemplate("dictExtract", extractDictTemplate, s)
}
func (s *Dictionary) RenderFormat() string { return "" }

func (s *Struct) RenderLiteral() string { return s.Name }
func (s *Struct) RenderName(isAbbreviate bool) string {
	if isAbbreviate {
		return s.AbbrName
	}
	return s.Name
}
func (s *Struct) RenderDefinition(isAbbreviate bool) string {
	return renderTemplate(
		"struct",
		structTemplate,
		struct {
			IsAbbreviate bool
			P            *Struct
		}{isAbbreviate, s})
}
func (s *Struct) RenderExtraction(vn, on string) string {
	return renderTemplate("structExtract", extractStructTemplate, s)
}

func (s *Struct) RenderValidate(name string) string {
	return renderTemplate(
		"structValidate",
		structValidateTemplate,
		struct {
			Name string
			P    *Struct
		}{name, s})
}
func (s *Struct) RenderFormat() string { return "" }

func (s *Struct) RenderTags(p property) string {
	tags := []string{fmt.Sprintf("json:\"%s\"", p.SourceName)}

	if p.Required || len(p.Enum) > 0 {
		validateTags := ""
		if p.Required {
			validateTags += "required"
		}
		if len(p.Enum) > 0 {
			if validateTags != "" {
				validateTags += ","
			}
			validateTags += "in("
			for i, el := range p.Enum {
				validateTags += el
				if i < len(p.Enum)-1 {
					validateTags += "|"
				}
			}
			validateTags += ")"
		}
		switch p.Reference.RenderFormat() {
		case "date-time", "date":
			if validateTags != "" {
				validateTags += ","
			}
			validateTags += "rfc3339"
		}

		tags = append(tags, fmt.Sprintf("valid:\"%s\"", validateTags))
	}

	return fmt.Sprintf("`%s`", strings.Join(tags, " "))
}

func (p *Param) RenderExtraction() string {
	return renderTemplate("paramExtract", paramTemplate, p)
}

func (ctx *Context) setProperty(schema *Schema, name, pname, rname, descPname string) property {

	var refName, desc string

	if rname != "" {
		refName = rname
		desc = rname
	} else {
		refName = ToCamelCase(true, pname, name)
		if descPname == "" {
			desc = refName
		} else {
			desc = fmt.Sprintf("%s.%s", descPname, ToCamelCase(true, name))
		}
	}

	p := property{
		Name:       ToCamelCase(true, name),
		SourceName: name,
		Enum:       schema.Enum,
	}

	switch schema.Type {
	case "object":
		if schema.AdditionalProperties == nil {
			ps := &Struct{
				Name:       refName,
				Properties: []property{},
				AbbrName:   ToAbbreviate(desc),
				Desc:       desc,
			}
			for n, s := range schema.Properties {
				p := ctx.setProperty(s, n, refName, getRefName(s.Ref), desc)
				for _, a := range schema.Required {
					if n == a {
						p.Required = true
					}
				}
				ps.Properties = append(ps.Properties, p)
			}
			p.Reference = ps
			ctx.References[refName] = p
		} else {
			p.Reference = &Dictionary{
				Name:      refName,
				ItemsType: ctx.setProperty(schema.AdditionalProperties, "", refName, getRefName(schema.AdditionalProperties.Ref), desc),
			}
		}
	case "string":
		p.Reference = &String{
			Default: schema.Default,
			Values:  schema.Enum,
			Format:  schema.Format,
		}
	case "integer":
		p.Reference = &Integer{}
	case "array":
		p.Reference = &Slice{
			Name:      getRefName(schema.Items.Ref),
			ItemsType: ctx.setProperty(schema.Items, "", refName, getRefName(schema.Items.Ref), desc),
		}
	case "number":
		p.Reference = &Number{}
	case "boolean":
		p.Reference = &Bool{}
	default:
		log.Fatalf("unsupported type %s", schema.Type)
	}
	return p
}

func (ctx *Context) getParams(ps []*Parameter, rb *RequestBody, opID string) []Param {
	inputs := []Param{}
	for _, p := range ps {
		inputs = append(inputs, newParam(p.In, p.Required, ctx.setProperty(p.Schema, p.ExternalName, opID, getRefName(p.Ref), ""))) //TODO:
	}
	if rb != nil {
		for k, mt := range rb.Content {
			if rb.Check(k) {
				inputs = append(inputs, newParam("body", rb.Required, ctx.setProperty(mt.Schema, "Request", opID, getRefName(rb.Ref), ""))) //TODO:
			}
		}
	}
	return inputs
}

func (ctx *Context) getResponses(rs map[string]*Response, opID string) []Param {
	inputs := []Param{}
	for c, response := range rs {
		code, err := strconv.Atoi(c)
		if err != nil {
			continue
		}
		if code >= http.StatusOK && code < http.StatusBadRequest {
			for k, mt := range response.Content {
				if response.Check(k) {
					inputs = append(inputs, newParam(c, true, ctx.setProperty(mt.Schema, "Response", opID, "", ""))) //TODO:
				}
			}
		}
	}
	return inputs
}

func (ot OperationType) String() string {
	return operationTypeValues[ot]
}

func (v functions) Len() int           { return len(v) }
func (v functions) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v functions) Less(i, j int) bool { return v[i].Name < v[j].Name }

func (v propertiesByLiteral) Len() int      { return len(v) }
func (v propertiesByLiteral) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v propertiesByLiteral) Less(i, j int) bool {
	return v[i].Reference.RenderLiteral() < v[j].Reference.RenderLiteral()
}

func (v propertiesByName) Len() int      { return len(v) }
func (v propertiesByName) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v propertiesByName) Less(i, j int) bool {
	return v[i].Name < v[j].Name
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

func ToAbbreviate(name string) string {
	return abbreviate(name)
}

func abbreviate(s string) string {
	abbr := make([]byte, 0, len(s))
	lex := make([]byte, 0, len(s))

	f := true
	for _, ch := range s {
		if ch == '.' {
			f = true
			continue
		}
		if f {
			abbr = append(abbr, byte(unicode.ToUpper(ch)))
			lex = make([]byte, 0, len(s))
			f = false
		} else {
			lex = append(lex, byte(ch))
		}
	}
	abbr = append(abbr, lex...)

	return string(abbr)
}

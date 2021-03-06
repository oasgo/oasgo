package main

import (
	"io"
	"os"
	"text/template"
)

const ClientTemplate = `
// Code generated by https://github.com/oasgo/oasgo. DO NOT EDIT.
// Source: {{ .Info.Title }} Version: {{ .Info.Version }}

// Package {{.PackageName}} is a generated OASGO package.

package {{.PackageName}}

{{ $cName :=  (printf "HTTP%sClient" (goName .Info.Title false) ) }}
{{ $iName := (printf (goName .Info.Title false)) }}

var _ {{ $iName }} = new({{ $cName }})

type (
    {{ $iName }} interface {
		{{- range $f := $.SortedFunctions }}
			{{$f.RenderSignature}}
		{{- end -}}
    }

    {{ $cName }} struct {
        URL *url.URL
        HTTP *http.Client
    }

	{{ range $r := $.SortedReferences }}
		{{$r.Reference.RenderDefinition $.IsAbbreviate}}
	{{ end }}
)

func New{{ $cName }} (host string) (*{{ $cName }}, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	return &{{$cName}}{
		URL: u,
    HTTP: &http.Client{},
	}, nil
}

{{- range $f := $.SortedFunctions }}
func (c *{{ $cName }}) {{$f.RenderSignature}} {
	{{- $f.RenderBody -}}
}
{{ end }}

func (c *{{ $cName }}) sendRequest(res interface{}, request *http.Request) (*http.Response, error){
	
resp, err := c.HTTP.Do(request)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return resp, nil
	}

	body := bytes.NewBuffer(make([]byte, 0))
	if r, ok := res.(*string); ok {
		var bs []byte
		if bs, err = ioutil.ReadAll(resp.Body); err == nil {
			*r = string(bs)
		}
		body = bytes.NewBuffer(bs)
	} else {
		err = json.NewDecoder(io.TeeReader(resp.Body, body)).Decode(res)
	}
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(body)
	
	return resp, err
}

`

func renderClient(s *Swagger, pn, dest string, isAbbreviate bool) {
	tmpl, err := template.New("client").Funcs(getFuncMap()).Parse(ClientTemplate)
	if err != nil {
		os.Stderr.WriteString("Parse tmpl error: " + err.Error())
		os.Exit(1)
	}

	c := Context{
		PackageName:  pn,
		Info:         s.Info,
		IsAbbreviate: isAbbreviate,
		References:   make(map[string]property),
		Functions:    []Function{},
	}

	for n, schema := range s.Components.Schemas {
		c.setProperty(schema, n, "", "", "")
	}
	for n, rb := range s.Components.RequestBodies {
		for k, mt := range rb.Content {
			if rb.Check(k) {
				c.setProperty(mt.Schema, n, "", "", "")
			}
		}
	}
	for n, response := range s.Components.Responses {
		for k, mt := range response.Content {
			if response.Check(k) {
				c.setProperty(mt.Schema, n, "", "", "")
			}
		}
	}

	for path, m := range s.Paths {
		if m.GET != nil {
			c.Functions = append(c.Functions, Function{
				Name:          ToCamelCase(true, m.GET.OperationID),
				Path:          path,
				OperationType: GET,
				Input:         c.getParams(m.GET.Parameters, nil, m.GET.OperationID),
				Output:        c.getResponses(m.GET.Responses, m.GET.OperationID),
			})
		}
		if m.POST != nil {
			c.Functions = append(c.Functions, Function{
				Name:          ToCamelCase(true, m.POST.OperationID),
				Path:          path,
				OperationType: POST,
				Input:         c.getParams(m.POST.Parameters, m.POST.RequestBody, m.POST.OperationID),
				Output:        c.getResponses(m.POST.Responses, m.POST.OperationID),
			})
		}
		if m.PUT != nil {
			c.Functions = append(c.Functions, Function{
				Name:          ToCamelCase(true, m.PUT.OperationID),
				Path:          path,
				OperationType: PUT,
				Input:         c.getParams(m.PUT.Parameters, m.PUT.RequestBody, m.PUT.OperationID),
				Output:        c.getResponses(m.PUT.Responses, m.PUT.OperationID),
			})
		}
		if m.PATCH != nil {
			c.Functions = append(c.Functions, Function{
				Name:          ToCamelCase(true, m.PATCH.OperationID),
				Path:          path,
				OperationType: PATCH,
				Input:         c.getParams(m.PATCH.Parameters, m.PATCH.RequestBody, m.PATCH.OperationID),
				Output:        c.getResponses(m.PATCH.Responses, m.PATCH.OperationID),
			})
		}
		if m.DELETE != nil {
			c.Functions = append(c.Functions, Function{
				Name:          ToCamelCase(true, m.DELETE.OperationID),
				Path:          path,
				OperationType: DELETE,
				Input:         c.getParams(m.DELETE.Parameters, nil, m.DELETE.OperationID),
				Output:        c.getResponses(m.DELETE.Responses, m.DELETE.OperationID),
			})
		}
	}

	var wr io.Writer = os.Stdout
	if dest != "" {
		f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			os.Stderr.WriteString("Cann't open destination file: " + err.Error())
			os.Exit(3)
		}
		wr = f
	}
	err = tmpl.Execute(wr, c)
	if err != nil {
		os.Stderr.WriteString("Execute tmpl error: " + err.Error())
		os.Exit(2)
	}
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"goName": func(name string, upper bool) string {
			return ToCamelCase(true, name)
		},
	}
}

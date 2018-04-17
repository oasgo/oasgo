package main

import (
	"os"
	"text/template"
)

const (
	DTOTemplate = `
/*
* This file autogenerated;
*
* DO NOT EDIT
* Generated by https://github.com/oasgo/oasgo
*/

package {{.PackageName}}
type (
	{{ range $r := .References }}
		{{$r.Reference.RenderDefinition}}
	{{ end }}
)
{{ range $r := .References }}
func (r *{{$r.Reference.Name}}) Validate() (bool, error) {
	return govalidator.ValidateStruct(r)
}
{{ end }}
`
)

func renderDTO(s *Swagger, pn string) {
	tmpl, err := template.New("dto").Parse(DTOTemplate)
	if err != nil {
		os.Stderr.WriteString("Parse tmpl error: " + err.Error())
		os.Exit(1)
	}

	c := Context{
		PackageName: pn,
		References:  make(map[string]property),
		Functions:   []Function{},
	}

	for n, schema := range s.Components.Schemas {
		c.setProperty(schema, n, "", "")
	}
	for n, rb := range s.Components.RequestBodies {
		for k, mt := range rb.Content {
			if rb.check(k) {
				c.setProperty(mt.Schema, n, "", "")
			}
		}
	}
	for n, response := range s.Components.Responses {
		for k, mt := range response.Content {
			if response.check(k) {
				c.setProperty(mt.Schema, n, "", "")
			}
		}
	}

	for _, m := range s.Paths {
		if m.GET != nil {
			c.Functions = append(c.Functions, Function{
				Name:   ToCamelCase(true, m.GET.OperationID),
				Input:  c.getParams(m.GET.Parameters, nil, m.GET.OperationID),
				Output: c.getResponses(m.GET.Responses, m.GET.OperationID),
			})
		}
		if m.POST != nil {
			c.Functions = append(c.Functions, Function{
				Name:   ToCamelCase(true, m.POST.OperationID),
				Input:  c.getParams(m.POST.Parameters, m.POST.RequestBody, m.POST.OperationID),
				Output: c.getResponses(m.POST.Responses, m.POST.OperationID),
			})
		}
		if m.PUT != nil {
			c.Functions = append(c.Functions, Function{
				Name:   ToCamelCase(true, m.PUT.OperationID),
				Input:  c.getParams(m.PUT.Parameters, m.PUT.RequestBody, m.PUT.OperationID),
				Output: c.getResponses(m.PUT.Responses, m.PUT.OperationID),
			})
		}
		if m.PATCH != nil {
			c.Functions = append(c.Functions, Function{
				Name:   ToCamelCase(true, m.PATCH.OperationID),
				Input:  c.getParams(m.PATCH.Parameters, m.PATCH.RequestBody, m.PATCH.OperationID),
				Output: c.getResponses(m.PATCH.Responses, m.PATCH.OperationID),
			})
		}
		if m.DELETE != nil {
			c.Functions = append(c.Functions, Function{
				Name:   ToCamelCase(true, m.DELETE.OperationID),
				Input:  c.getParams(m.DELETE.Parameters, nil, m.DELETE.OperationID),
				Output: c.getResponses(m.DELETE.Responses, m.DELETE.OperationID),
			})
		}
	}

	err = tmpl.Execute(os.Stdout, c)
	if err != nil {
		os.Stderr.WriteString("Execute tmpl error: " + err.Error())
		os.Exit(2)
	}
}

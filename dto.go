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
		c.setProperty(schema, n, "")
	}
	err = tmpl.Execute(os.Stdout, c)
	if err != nil {
		os.Stderr.WriteString("Execute tmpl error: " + err.Error())
		os.Exit(2)
	}
}

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

var templates = []string{
	"templates/package.tmpl",
	"templates/signature.tmpl",
	"templates/struct.tmpl",
	"templates/array.tmpl",
}

var predefinedTypes = map[string]string{
	"string":  "string",
	"integer": "int",
}

// Swagger https://swagger.io/specification/
type Swagger struct {
	Info       Info
	OpenAPI    string
	Servers    []Server
	Paths      map[string]PathItem
	Components Components
}

// Info https://swagger.io/specification/#infoObject
type Info struct {
	Title   string
	Version string
}

// PathItem https://swagger.io/specification/#pathItemObject
type PathItem struct {
	GET    *Operation
	POST   *Operation
	PATCH  *Operation
	PUT    *Operation
	DELETE *Operation
}

// Operation https://swagger.io/specification/#operationObject
type Operation struct {
	OperationID string `yaml:"operationId"`
	Summary     string
	Description string
	Parameters  []Parameter
	Responses   map[string]*Response
}

// Parameter https://swagger.io/specification/#parameterObject
type Parameter struct {
	Name     string
	In       string
	Required bool
	Schema   *Schema
}

// Response https://swagger.io/specification/#responseObject
type Response struct {
	Description string
	Headers     map[string]*Header
	Content     map[string]*MediaType
	Links       map[string]*Link
}

// Server https://swagger.io/specification/#serverObject
type Server struct {
	URL string
}

// Components https://swagger.io/specification/#componentsObject
type Components struct {
	Schemas map[string]*Schema
}

// Schema https://swagger.io/specification/#schemaObject
type Schema struct {
	Name       string
	Ref        string `yaml:"$ref"`
	Type       string
	Format     string
	Required   []string
	Properties map[string]*Schema
	Items      *Schema
}

// Header https://swagger.io/specification/#headerObject
type Header struct {
	Name        string
	Description string
}

// MediaType https://swagger.io/specification/#mediaTypeObject
type MediaType struct {
	Schema *Schema
}

// Link https://swagger.io/specification/#linkObject
type Link struct {
	operationRef string
	operationID  string `yaml:"operationId"`
	description  string
}

func newSwagger(data []byte) (*Swagger, error) {
	swagger := Swagger{}

	if err := yaml.Unmarshal(data, &swagger); err != nil {
		return nil, err
	}

	// Resolve Schema references
	Inspect(swagger, func(n interface{}) bool {
		if schemaDest, ok := n.(*Schema); ok && schemaDest.Ref != "" {
			refName := getRefName(schemaDest.Ref)

			Inspect(swagger, func(n interface{}) bool {
				if schemaSource, ok := n.(*Schema); ok && schemaSource.Name == refName {
					ref := schemaDest.Ref
					*schemaDest = *schemaSource
					schemaDest.Ref = ref

					return false
				}

				return true
			})
		}

		return true
	})

	return &swagger, nil
}

// GetMethodsMap converts PathItem fields to map but without nil Operations.
func (p PathItem) GetMethodsMap() map[string]*Operation {
	m := make(map[string]*Operation)
	if p.GET != nil {
		m["GET"] = p.GET
	}
	if p.POST != nil {
		m["POST"] = p.POST
	}
	if p.PUT != nil {
		m["PUT"] = p.PUT
	}
	if p.PATCH != nil {
		m["PATCH"] = p.PATCH
	}
	if p.DELETE != nil {
		m["DELETE"] = p.DELETE
	}
	return m
}

// GetResult returns Schema with good status code.
func (o *Operation) GetResult() (r *Schema) {
	for c, r := range o.Responses {
		code, err := strconv.Atoi(c)
		if err != nil {
			continue
		}
		if code >= http.StatusOK && code < http.StatusBadRequest {
			return r.Content["application/json"].Schema
		}
	}
	return
}

func parse(path string) *Swagger {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		flag.PrintDefaults()
		fmt.Println(err)
		os.Exit(1)
	}

	s, err := newSwagger(data)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return s
}

// UnmarshalYAML defines default Type for Schema struct.
func (s *Schema) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rs Schema

	r := rs{Type: "object"}

	if err := unmarshal(&r); err != nil {
		return err
	}

	*s = Schema(r)

	return nil
}

// UnmarshalYAML fills Name for Schemas struct from map keys.
func (c *Components) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rc Components

	r := rc{}

	if err := unmarshal(&r); err != nil {
		return err
	}

	for k, v := range r.Schemas {
		v.Name = k
	}

	*c = Components(r)

	return nil
}

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	yaml "gopkg.in/yaml.v2"
	"net/url"
)

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
	RequestBody *RequestBody `yaml:"requestBody"`
	Parameters  []*Parameter
	Responses   map[string]*Response
}

// RequestBody https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/versions/3.0.0.md#requestBodyObject
type RequestBody struct {
	Description string
	Required    bool
	Content     map[string]*MediaType
	Ref         string `yaml:"$ref"`
}

// Parameter https://swagger.io/specification/#parameterObject
type Parameter struct {
	Name         string
	ExternalName string
	In           string
	Required     bool
	Schema       *Schema
	Ref          string `yaml:"$ref"`
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
	Schemas       map[string]*Schema
	Parameters    map[string]*Parameter
	RequestBodies map[string]*RequestBody `yaml:"requestBodies"`
	Responses     map[string]*Response    `yaml:"responses"`
}

// Schema https://swagger.io/specification/#schemaObject
type Schema struct {
	Name                 string
	Ref                  string `yaml:"$ref"`
	Type                 string
	Format               string
	Required             []string
	Properties           map[string]*Schema
	Items                *Schema
	Parent               *Schema
	AdditionalProperties *Schema             `yaml:"additionalProperties"`
	Enum                 []string            `yaml:"enum"`
	Default              string              `yaml:"default"`
	ExtensionTags        map[string][]string `yaml:"x-oasgo-tags"`
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
		if schemaDest, ok := n.(*Schema); ok && schemaDest != nil && schemaDest.Ref != "" {
			refName := getRefName(schemaDest.Ref)
			Inspect(swagger, func(n interface{}) bool {
				if schemaSource, ok := n.(*Schema); ok && schemaSource != nil && schemaSource.Name == refName {
					ref := schemaDest.Ref
					*schemaDest = *schemaSource
					schemaDest.Ref = ref

					return false
				}
				return true
			})
		}

		if parameterDest, ok := n.(*Parameter); ok {
			if parameterDest.Ref != "" {
				refName := getRefName(parameterDest.Ref)
				Inspect(swagger, func(n interface{}) bool {
					if parameterSource, ok := n.(*Parameter); ok && parameterSource.Name == refName {
						ref := parameterDest.Ref
						*parameterDest = *parameterSource
						parameterDest.Ref = ref
						return false
					}

					return true
				})
			} else if parameterDest.ExternalName == "" {
				parameterDest.ExternalName = parameterDest.Name
			}
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
	for k, v := range r.Parameters {
		v.ExternalName = v.Name
		v.Name = k
	}

	*c = Components(r)

	return nil
}

func (rb *RequestBody) Check(key string) bool {
	return check([]string{"application/json"}, key)
}

func (rb *Response) Check(key string) bool {
	return check([]string{"application/json"}, key)
}

// Inspect calls visitor function on (almost) every node within Swagger struct.
func Inspect(node interface{}, visitor func(i interface{}) bool) {
	if ok := visitor(node); !ok {
		return
	}

	switch n := node.(type) {
	case Swagger:
		Inspect(n.Components, visitor)

		for _, v := range n.Paths {
			Inspect(v, visitor)
		}
	case PathItem:
		if n.GET != nil {
			Inspect(n.GET, visitor)
		}
		if n.POST != nil {
			Inspect(n.POST, visitor)
		}
		if n.PATCH != nil {
			Inspect(n.PATCH, visitor)
		}
		if n.PUT != nil {
			Inspect(n.PUT, visitor)
		}
		if n.DELETE != nil {
			Inspect(n.DELETE, visitor)
		}
	case *Operation:
		for _, v := range n.Parameters {
			Inspect(v, visitor)
		}
		for _, v := range n.Responses {
			Inspect(v, visitor)
		}
		if n.RequestBody != nil {
			Inspect(n.RequestBody, visitor)
		}
	case *Parameter:
		if n.Schema != nil {
			Inspect(n.Schema, visitor)
		}
	case *Response:
		for _, v := range n.Content {
			Inspect(v, visitor)
		}
	case *MediaType:
		if n.Schema != nil {
			Inspect(n.Schema, visitor)
		}
	case Components:
		for _, v := range n.Schemas {
			Inspect(v, visitor)
		}
		for _, v := range n.Parameters {
			Inspect(v, visitor)
		}
		for _, v := range n.RequestBodies {
			Inspect(v, visitor)
		}
		for _, v := range n.Responses {
			Inspect(v, visitor)
		}
	case *Schema:
		if n.Items != nil {
			n.Items.Parent = n
			Inspect(n.Items, visitor)
		}
		for _, v := range n.Properties {
			v.Parent = n
			Inspect(v, visitor)
		}
	case *RequestBody:
		for _, v := range n.Content {
			Inspect(v, visitor)
		}
	}
}

func parse(path string) *Swagger {
	data, err := readFile(path)
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

func check(availableKeys []string, key string) bool {
	for _, el := range availableKeys {
		if key == el {
			return true
		}
	}
	return false
}

// readFile Read the file by URL, if the path is a reference, else from the local file
func readFile(path string) ([]byte, error) {
	if _, err := url.ParseRequestURI(path); err!= nil{
		return ioutil.ReadFile(path)
	}

	r, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return ioutil.ReadAll(r.Body)
}

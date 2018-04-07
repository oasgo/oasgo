package main

// Inspect calls visitor function on (almost) every node within Swagger struct.
//
//   var schema *Schema
//   Inspect(swagger, func(node interface{}) bool {
//     if s, ok := node.(*Schema); ok && s.Name == "Pet" {
//       schema = s
//       return false
//     }
//     return true
//   })
//   fmt.Println(schema)
//
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
	case *Schema:
		if n.Items != nil {
			n.Items.Parent = n
			Inspect(n.Items, visitor)
		}
		for _, v := range n.Properties {
			v.Parent = n
			Inspect(v, visitor)
		}
	}
}

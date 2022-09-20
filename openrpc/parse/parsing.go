package parse

import (
	"strings"

	"gfx.cafe/open/jrpc/openrpc/types"
	"gfx.cafe/open/jrpc/openrpc/util"
	"github.com/go-openapi/spec"
)

type GoStruct struct {
	Name        string
	Description string

	Fields map[string]string
}

type GoField struct {
	Name     string
	Type     *GoType
	Optional bool

	Description string
}

type GoType struct {
	Name string
	Type string

	Description string
}

type GoGlobal struct {
	Structs map[string]*GoStruct
	Types   map[string]*GoType
	Methods map[string]*GoMethod
}

type GoMethod struct {
	Input  string
	Output string
	Name   string
}

func (g *GoGlobal) ReadOpenRPC(s types.OpenRPCSpec1) {
	for k, v := range s.Components.Schemas {
		g.AddSchema(k, v)
	}
	for _, v := range s.Methods {
		g.AddMethod(v.Name, v)
	}
}

func (g *GoGlobal) AddSchema(name string, m spec.Schema) {
	if len(m.Type) > 0 {
		switch m.Type[0] {
		case "string":
		case "number":
		case "boolean":
		case "object":
		case "array":
		default:
		}
		return
	}
	if len(m.OneOf) > 0 {
		//todo: handle
		return
	}
}

func (g *GoGlobal) AddMethod(name string, m types.Method) {
	inStruct := "Params" + util.CamelCase(name)
	outStruct := ""
	if m.Result != nil {
		outStruct = "Result" + util.CamelCase(name)
	}
	g.Methods[name] = &GoMethod{
		Input:  inStruct,
		Output: outStruct,
		Name:   util.CamelCase(name),
	}
}

func (g *GoGlobal) AddStruct(name string, m spec.Schema) {
	s := &GoStruct{}
	g.Structs[name] = s
	s.Name = name
	s.Description = m.Description
	s.Fields = map[string]string{}
	for k, v := range m.Properties {
		g.AddSchema(k, v)
		if ref := v.Ref.GetPointer(); ref != nil {
			refstr := strings.ToTitle(ref.String())
			s.Fields[k] = refstr
		} else {
			if len(v.Type) > 0 {
				refstr := v.Type[0]
				s.Fields[k] = refstr
			}
		}
	}
}

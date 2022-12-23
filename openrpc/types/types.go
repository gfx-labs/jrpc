package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

type Server struct {
	Name        string                    `json:"name"`
	URL         string                    `json:"url"`
	Summary     string                    `json:"summary,omitempty"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description,omitempty"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Items []Schema

func (I *Items) UnmarshalJSON(b []byte) error {
	switch b[0] {
	case '{':
		*I = []Schema{
			{},
		}
		return json.Unmarshal(b, &(*I)[0])
	case '[':
		return json.Unmarshal(b, (*[]Schema)(I))
	default:
		return fmt.Errorf("expected array or object")
	}
}

var _ json.Unmarshaler = (*Items)(nil)

type Schema struct {
	Ref        string            `json:"$ref,omitempty"`
	Type       string            `json:"type,omitempty"`
	Title      string            `json:"title,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Items      Items             `json:"items,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	OneOf      []Schema          `json:"oneOf,omitempty"`
	AnyOf      []Schema          `json:"anyOf,omitempty"`
	AllOf      []Schema          `json:"allOf,omitempty"`
	Enum       []string          `json:"enum,omitempty"`
	Pattern    string            `json:"pattern,omitempty"`
}

type Param struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

type Result struct {
	Name   string `json:"name"`
	Schema Schema `json:"schema"`
}

type Method struct {
	Name     string           `json:"name"`
	Tags     []Tag            `json:"tags,omitempty"`
	Summary  string           `json:"summary"`
	Params   []Param          `json:"params"`
	Result   Result           `json:"result"`
	Examples []ExamplePairing `json:"examples,omitempty"`
}

type Tag struct {
	Ref          string                `json:"$ref,omitempty"`
	Name         string                `json:"name"`
	Summary      string                `json:"summary,omitempty"`
	Description  string                `json:"description,omitempty"`
	ExternalDocs ExternalDocumentation `json:"externalDocs,omitempty"`
}

type ExternalDocumentation struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

type ExamplePairing struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	Params      []Example `json:"params"`
	Result      Example   `json:"result"`
}

type Example struct {
	Ref         string `json:"$ref,omitempty"`
	Name        string `json:"name"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	Value       any    `json:"value"`
}

func (m *Method) Namespace() string {
	splt := strings.Split(m.Name, "_")
	if len(splt) > 1 {
		return splt[0]
	}
	return ""
}

func (m *Method) MethodName() string {
	splt := strings.Split(m.Name, "_")
	if len(splt) > 1 {
		return splt[1]
	}
	return splt[0]
}

type OpenRPC struct {
	Package    string   `json:"-"`
	Version    string   `json:"openrpc"`
	Info       Info     `json:"info"`
	Methods    []Method `json:"methods"`
	Servers    []Server `json:"servers"`
	Components struct {
		Schemas map[string]Schema `json:"schemas"`
	} `json:"components"`
}

func NewOpenRPCSpec1() *OpenRPC {
	return &OpenRPC{
		Package: "main",
		Version: "1.0.0",
		Info: Info{
			Title:   "gfx.cafe/open/jrpc/openrpc",
			Version: "0.0.0",
		},
		Servers: make([]Server, 0),
		Methods: make([]Method, 0),
	}
}

func (o *OpenRPC) AddSchemas(pth string) error {
	dr, err := os.ReadDir(pth)
	if err != nil {
		return err
	}
	for _, v := range dr {
		if v.IsDir() {
			if err := o.AddSchemas(path.Join(pth, v.Name())); err != nil {
				return err
			}
		} else {
			if err := o.AddSchema(path.Join(pth, v.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *OpenRPC) AddSchema(pth string) error {
	schem := map[string]Schema{}
	bts, err := os.ReadFile(pth)
	if err != nil {
		return err
	}
	ext := filepath.Ext(path.Base(pth))
	switch ext {
	case ".json":
		err = json.Unmarshal(bts, &schem)
	case ".yml", ".yaml":
		err = yaml.Unmarshal(bts, &schem)
	}
	if err != nil {
		return err
	}
	if o.Components.Schemas == nil {
		o.Components.Schemas = map[string]Schema{}
	}
	for k, v := range schem {
		o.Components.Schemas[k] = v
	}
	return nil
}

func (o *OpenRPC) AddMethods(pth string) error {
	dr, err := os.ReadDir(pth)
	if err != nil {
		return err
	}
	for _, v := range dr {
		if v.IsDir() {
			if err := o.AddMethods(path.Join(pth, v.Name())); err != nil {
				return err
			}
		} else {
			if err := o.AddMethod(path.Join(pth, v.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *OpenRPC) AddMethod(pth string) error {
	var meth []Method
	bts, err := os.ReadFile(pth)
	if err != nil {
		return err
	}
	switch filepath.Ext(path.Base(pth)) {
	case ".json":
		err = json.Unmarshal(bts, &meth)
	case ".yml", ".yaml":
		err = yaml.Unmarshal(bts, &meth)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	o.Methods = append(o.Methods, meth...)
	return nil
}

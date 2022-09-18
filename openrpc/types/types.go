package types

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/yaml"
)

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description,omitempty"`
}

type Server struct {
	Name        string                    `json:"name"`
	URL         string                    `json:"url"`
	Summary     string                    `json:"summary,omitempty"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

type Tag struct {
	Name         string        `json:"name"`
	Summary      string        `json:"summary,omitempty"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

type Content struct {
	Name        string      `json:"name"`
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Deprecated  bool        `json:"deprecated,omitempty"`
	Schema      spec.Schema `json:"schema"`
}

type ContentDescriptor struct {
	Content
}

func (cd *ContentDescriptor) UnmarshalJSON(data []byte) error {
	cont := new(Content)
	err := json.Unmarshal(data, cont)
	if err != nil {
		return err
	}
	cd.Content = *cont

	params := make(map[string]interface{})
	err = json.Unmarshal(data, &params)
	if err != nil {
		return err
	}

	if _, ok := params["$ref"]; ok {
		sch := new(spec.Schema)
		err = json.Unmarshal(data, sch)
		if err != nil {
			return err
		}
		cd.Schema = *sch
	}

	return nil

}

// https://www.jsonrpc.org/specification#error_object
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Link struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Summary     string                 `json:"summary"`
	Method      string                 `json:"method"`
	Params      map[string]interface{} `json:"params"`
	Server      Server                 `json:"server"`
}

type Example struct {
	Name          string      `json:"name"`
	Summary       string      `json:"summary"`
	Description   string      `json:"description"`
	Value         interface{} `json:"value"`
	ExternalValue string      `json:"externalValue"`
}

type ExamplePairing struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Summary     string    `json:"summary"`
	Params      []Example `json:"params"`
	Result      Example   `json:"result"`
}

type Method struct {
	Name           string               `json:"name"`
	Tags           []Tag                `json:"tags,omitempty"`
	Summary        string               `json:"summary,omitempty"`
	Description    string               `json:"description,omitempty"`
	ExternalDocs   *ExternalDocs        `json:"externalDocs,omitempty"`
	Params         []*ContentDescriptor `json:"params"`
	Result         *ContentDescriptor   `json:"result"`
	Deprecated     bool                 `json:"deprecated,omitempty"`
	Servers        []Server             `json:"servers,omitempty"`
	Errors         []Error              `json:"errors,omitempty"`
	Links          []Link               `json:"links,omitempty"`
	ParamStructure string               `json:"paramStructure,omitempty"`
	Examples       []ExamplePairing     `json:"examples,omitempty"`
}

type Components struct {
	ContentDescriptors    map[string]*ContentDescriptor `json:"contentDescriptors,omitempty"`
	Schemas               map[string]spec.Schema        `json:"schemas,omitempty"`
	Examples              map[string]Example            `json:"examples,omitempty"`
	Links                 map[string]Link               `json:"links,omitempty"`
	Errors                map[string]Error              `json:"errors,omitempty"`
	ExamplePairingObjects map[string]ExamplePairing     `json:"examplePairingObjects,omitempty"`
	Tags                  map[string]Tag                `json:"tags,omitempty"`
}

type OpenRPCSpec1 struct {
	OpenRPC      string        `json:"openrpc"`
	Info         Info          `json:"info"`
	Servers      []Server      `json:"servers"`
	Methods      []Method      `json:"methods"`
	Components   Components    `json:"components"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`

	Objects *ObjectMap `json:"-"`
}

func NewOpenRPCSpec1() *OpenRPCSpec1 {
	return &OpenRPCSpec1{
		OpenRPC: "1.0.0",
		Info: Info{
			Title:   "gfx.cafe/open/jrpc/openrpc",
			Version: "0.0.0",
		},
		Servers: make([]Server, 0),
		Methods: make([]Method, 0),

		Objects: NewObjectMap(),
	}
}

func (o *OpenRPCSpec1) AddSchemas(pth string) error {
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

func (o *OpenRPCSpec1) AddSchema(pth string) error {
	schem := map[string]spec.Schema{}
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
		o.Components.Schemas = map[string]spec.Schema{}
	}
	for k, v := range schem {
		o.Components.Schemas[k] = v
	}
	return nil
}

func (o *OpenRPCSpec1) AddMethods(pth string) error {
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

func (o *OpenRPCSpec1) AddMethod(pth string) error {
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

package types

import (
	"encoding/json"
	"fmt"
)

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
	return nil
}

var _ json.Unmarshaler = (*Items)(nil)

type Schema struct {
	Ref        string            `json:"$ref,omitempty"`
	Type       string            `json:"type,omitempty"`
	Title      string            `json:"title"`
	Required   []string          `json:"required,omitempty"`
	Items      Items             `json:"items,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	OneOf      []Schema          `json:"oneOf,omitempty"`
	AnyOf      []Schema          `json:"anyOf,omitempty"`
	AllOf      []Schema          `json:"allOf,omitempty"`
	Enum       []string          `json:"enum,omitempty"`
	Pattern    string            `json:"pattern"`
}

type Param struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Schema   Schema `json:"schema"`
}

type Result struct {
	Name   string `json:"name"`
	Schema Schema `json:"schema"`
}

type Method struct {
	Name    string  `json:"name"`
	Summary string  `json:"summary"`
	Params  []Param `json:"params"`
	Result  Result  `json:"result"`
}

type OpenRPC struct {
	Package    string   `json:"package"`
	Version    string   `json:"openrpc"`
	Info       Info     `json:"info"`
	Methods    []Method `json:"methods"`
	Components struct {
		Schemas map[string]Schema `json:"schemas"`
	} `json:"components"`
}

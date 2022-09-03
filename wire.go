package jrpc

import (
	"encoding/json"
	"fmt"
)

// Version represents a JSON-RPC version.
const Version = "2.0"

// version is a special 0 sized struct that encodes as the jsonrpc version tag.
//
// It will fail during decode if it is not the correct version tag in the stream.
type version struct{}

// compile time check whether the version implements a json.Marshaler and json.Unmarshaler interfaces.
var (
	_ json.Marshaler   = (*version)(nil)
	_ json.Unmarshaler = (*version)(nil)
)

// MarshalJSON implements json.Marshaler.
func (version) MarshalJSON() ([]byte, error) {
	return json.Marshal(Version)
}

// UnmarshalJSON implements json.Unmarshaler.
func (version) UnmarshalJSON(data []byte) error {
	version := ""
	if err := json.Unmarshal(data, &version); err != nil {
		return fmt.Errorf("failed to Unmarshal: %w", err)
	}
	if version != Version {
		return fmt.Errorf("invalid RPC version %v", version)
	}
	return nil
}

// ID is a Request identifier.
//
// Only one of either the Name or Number members will be set, using the
// number form if the Name is the empty string.
// alternatively, ID can be null
type ID struct {
	name   string
	number int32

	null bool
}

// compile time check whether the ID implements a fmt.Formatter, json.Marshaler and json.Unmarshaler interfaces.
var (
	_ fmt.Formatter    = (*ID)(nil)
	_ json.Marshaler   = (*ID)(nil)
	_ json.Unmarshaler = (*ID)(nil)
)

// NewNumberID returns a new number request ID.
func NewNumberID(v int32) ID { return *NewNumberIDPtr(v) }

// NewStringID returns a new string request ID.
func NewStringID(v string) ID { return *NewStringIDPtr(v) }

// NewStringID returns a new string request ID.
func NewNullID() ID { return *NewNullIDPtr() }

func NewNumberIDPtr(v int32) *ID  { return &ID{number: v} }
func NewStringIDPtr(v string) *ID { return &ID{name: v} }
func NewNullIDPtr() *ID           { return &ID{null: true} }

// Format writes the ID to the formatter.
//
// If the rune is q the representation is non ambiguous,
// string forms are quoted, number forms are preceded by a #.
func (id ID) Format(f fmt.State, r rune) {
	numF, strF := `%d`, `%s`
	if r == 'q' {
		numF, strF = `#%d`, `%q`
	}

	switch {
	case id.name != "":
		fmt.Fprintf(f, strF, id.name)
	default:
		fmt.Fprintf(f, numF, id.number)
	}
	id.null = false
}

// get the raw message
func (id *ID) RawMessage() json.RawMessage {
	if id.null {
		return null
	}
	if id.name != "" {
		ans, err := json.Marshal(id.name)
		if err == nil {
			return ans
		}
	}
	ans, err := json.Marshal(id.number)
	if err == nil {
		return ans
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (id *ID) MarshalJSON() ([]byte, error) {
	if id == nil {
		return null, nil
	}
	if id.null {
		return null, nil
	}
	if id.name != "" {
		return json.Marshal(id.name)
	}
	return json.Marshal(id.number)
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *ID) UnmarshalJSON(data []byte) error {
	*id = ID{}
	if err := json.Unmarshal(data, &id.number); err == nil {
		return nil
	}
	if err := json.Unmarshal(data, &id.name); err == nil {
		return nil
	}
	id.null = true
	return nil
}

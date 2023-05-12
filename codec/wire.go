package codec

import (
	"fmt"
	"strconv"

	json "github.com/goccy/go-json"
)

// Version represents a JSON-RPC version.
const VersionString = "2.0"

// version is a special 0 sized struct that encodes as the jsonrpc version tag.
//
// It will fail during decode if it is not the correct version tag in the stream.
type Version struct{}

// compile time check whether the version implements a json.Marshaler and json.Unmarshaler interfaces.
var (
	_ json.Marshaler   = (*Version)(nil)
	_ json.Unmarshaler = (*Version)(nil)
)

// MarshalJSON implements json.Marshaler.
func (Version) MarshalJSON() ([]byte, error) {
	return []byte(`"` + VersionString + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (Version) UnmarshalJSON(data []byte) error {
	version := ""
	if err := json.Unmarshal(data, &version); err != nil {
		return fmt.Errorf("failed to Unmarshal: %w", err)
	}
	if version != VersionString {
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
	number int64

	null bool

	empty bool
}

// compile time check whether the ID implements a fmt.Formatter, json.Marshaler and json.Unmarshaler interfaces.
var (
	_ fmt.Formatter    = (*ID)(nil)
	_ json.Marshaler   = (*ID)(nil)
	_ json.Unmarshaler = (*ID)(nil)
)

// NewNumberID returns a new number request ID.
func NewNumberID(v int64) ID { return *NewNumberIDPtr(v) }

// NewStringID returns a new string request ID.
func NewStringID(v string) ID { return *NewStringIDPtr(v) }

// NewStringID returns a new string request ID.
func NewNullID() ID { return *NewNullIDPtr() }

func NewNumberIDPtr(v int64) *ID { return &ID{number: v} }
func NewStringIDPtr(v string) *ID {
	if v == "" {
		return nil
	}
	return &ID{name: v}
}
func NewNullIDPtr() *ID { return &ID{null: true} }

func (id *ID) Number() int {
	if id == nil {
		return 0
	}
	if id.number == 0 {
		ans, _ := strconv.Atoi(id.name)
		return ans
	}
	return int(id.number)
}

// Format writes the ID to the formatter.
//
// If the rune is q the representation is non ambiguous,
// string forms are quoted, number forms are preceded by a #.
func (id *ID) Format(f fmt.State, r rune) {
	numF, strF := `%d`, `%s`
	if r == 'q' {
		numF, strF = `#%d`, `%q`
	}

	id.null = false
	switch {
	case id.name != "":
		fmt.Fprintf(f, strF, id.name)
	default:
		fmt.Fprintf(f, numF, id.number)
	}
}
func (id *ID) IsNull() bool {
	if id == nil {
		return true
	}
	return id.null
}

// get the raw message
func (id *ID) RawMessage() json.RawMessage {
	if id.empty {
		return nil
	}
	if id == nil {
		return Null
	}
	if id.null {
		return Null
	}
	if id.name != "" {
		return json.RawMessage(`"` + id.name + `"`)
	}
	return strconv.AppendInt(make([]byte, 0, 8), id.number, 10)
}

// MarshalJSON implements json.Marshaler.
func (id *ID) MarshalJSON() ([]byte, error) {
	return id.RawMessage(), nil
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

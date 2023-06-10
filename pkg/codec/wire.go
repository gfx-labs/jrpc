package codec

import (
	"bytes"
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
// alternatively, ID can be null
type ID json.RawMessage

func (i *ID) Format(f fmt.State, verb rune) {
	if i == nil {
		f.Write(Null)
		return
	}
	f.Write(*i)
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

func NewNumberIDPtr(v int64) *ID {
	o := ID(strconv.Itoa(int(v)))
	return &o
}
func NewStringIDPtr(v string) *ID {
	if v == "" {
		return nil
	}
	o := ID(`"` + v + `"`)
	return &o
}
func NewNullIDPtr() *ID {
	o := ID("null")
	return &o
}

func (id *ID) Number() int {
	if id == nil {
		return 0
	}
	ans, _ := strconv.Atoi(string(bytes.Trim(*id, `"'`)))
	return ans
}

func (id *ID) IsNull() bool {
	if id == nil {
		return false
	}
	return len(*id) == 4 &&
		(*id)[0] == 'n' &&
		(*id)[1] == 'u' &&
		(*id)[2] == 'l' &&
		(*id)[3] == 'l'
}

// get the raw message
func (id *ID) RawMessage() json.RawMessage {
	if id == nil {
		return Null
	}
	return json.RawMessage(*id)
}

// MarshalJSON implements json.Marshaler.
func (id *ID) MarshalJSON() ([]byte, error) {
	return id.RawMessage(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *ID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	*id = data
	// now validate
	if id.IsNull() {
		return nil
	}
	// it has to be a string or number
	var num int
	err := json.Unmarshal(data, &num)
	if err == nil {
		return nil
	}
	var str string
	err = json.Unmarshal(data, &str)
	if err == nil {
		return nil
	}
	*id = NewNullID()
	return fmt.Errorf("invalid id")
}

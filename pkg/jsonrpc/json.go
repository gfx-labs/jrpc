package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/go-faster/jx"
)

const NullString = "null"

var Null = json.RawMessage(NullString)

func NewNull() json.RawMessage {
	return json.RawMessage("null")
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type Message struct {
	ID          *ID                        `json:"id,omitempty"`
	Method      string                     `json:"method,omitempty"`
	Params      json.RawMessage            `json:"params,omitempty"`
	Error       error                      `json:"error,omitempty"`
	ExtraFields map[string]json.RawMessage `json:"-"`

	Result io.ReadCloser `json:"result,omitempty"`
}

func NewStringReader(x string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(x))
}

func MarshalMessage(m *Message, enc *jx.Encoder) (err error) {
	// use encoder
	fail := enc.Obj(func(e *jx.Encoder) {
		e.Field("jsonrpc", func(e *jx.Encoder) {
			e.Str("2.0")
		})
		if m.ID != nil {
			e.Field("id", func(e *jx.Encoder) {
				e.Raw(m.ID.RawMessage())
			})
		}
		if m.Method != "" {
			e.Field("method", func(e *jx.Encoder) {
				e.Str(m.Method)
			})
		}
		if m.Error != nil {
			e.Field("error", func(e *jx.Encoder) {
				e.Raw(MarshalError(m.Error))
			})
			return
		}
		if len(m.Params) != 0 {
			e.Field("params", func(e *jx.Encoder) {
				e.Raw(m.Params)
			})
		}
		if m.Result != nil && err == nil {
			e.Field("result", func(e *jx.Encoder) {
				var n int64
				n, err = io.Copy(e, m.Result)
				if n == 0 {
					e.Null()
				}
			})
		}
		if m.ExtraFields != nil {
			for k, v := range m.ExtraFields {
				e.Field(k, func(e *jx.Encoder) {
					e.Raw(v)
				})
			}
		}
	})
	if err != nil {
		return err
	}
	if fail {
		return fmt.Errorf("jx encoding error")
	}
	// output
	return nil
}

// parseID parses an ID from the decoder and returns a properly typed ID pointer
func parseID(d *jx.Decoder) (*ID, error) {
	switch d.Next() {
	case jx.Null:
		if err := d.Null(); err != nil {
			return nil, err
		}
		return NewNullIDPtr(), nil
	case jx.Number:
		num, err := d.Num()
		if err != nil {
			return nil, err
		}
		// Convert to int64 and create ID
		val, err := num.Int64()
		if err != nil {
			return nil, err
		}
		return NewNumberIDPtr(val), nil
	case jx.String:
		str, err := d.Str()
		if err != nil {
			return nil, err
		}
		return NewStringIDPtr(str), nil
	default:
		return nil, fmt.Errorf("invalid id type")
	}
}

// ParseIDBytes parses an ID from raw JSON bytes using jx decoder
func ParseIDBytes(data []byte) (*ID, error) {
	dec := jx.GetDecoder()
	defer jx.PutDecoder(dec)
	dec.ResetBytes(data)
	return parseID(dec)
}

// parseError parses a JSON-RPC error from the decoder
func parseError(d *jx.Decoder) (*JsonError, error) {
	je := &JsonError{}
	err := d.Obj(func(ed *jx.Decoder, key string) error {
		switch key {
		case "code":
			code, err := ed.Int()
			if err != nil {
				return err
			}
			je.Code = code
		case "message":
			msg, err := ed.Str()
			if err != nil {
				return err
			}
			je.Message = msg
		case "data":
			// For data, we need to keep it as raw for flexibility
			raw, err := ed.Raw()
			if err != nil {
				return err
			}
			// Unmarshal to any type
			err = json.Unmarshal(raw, &je.Data)
			if err != nil {
				return err
			}
		default:
			// Skip unknown fields
			if err := ed.Skip(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return je, nil
}

// ParseErrorBytes parses a JSON-RPC error from raw JSON bytes using jx decoder
func ParseErrorBytes(data []byte) (*JsonError, error) {
	dec := jx.GetDecoder()
	defer jx.PutDecoder(dec)
	dec.ResetBytes(data)
	return parseError(dec)
}

func UnmarshalMessage(m *Message, dec *jx.Decoder) error {
	err := dec.Obj(func(d *jx.Decoder, key string) (err error) {
		switch key {
		default:
			raw, err := d.Raw()
			if err != nil {
				return err
			}
			if m.ExtraFields == nil {
				m.ExtraFields = make(map[string]json.RawMessage)
			}
			// NOTE: we clone these.
			m.ExtraFields[key] = json.RawMessage(slices.Clone(raw))
		case "jsonrpc":
			value, err := d.Str()
			if err != nil {
				return err
			}
			if value != VersionString {
				return NewInvalidRequestError("Invalid Version")
			}
		case "id":
			id, err := parseID(d)
			if err != nil {
				return err
			}
			m.ID = id
		case "method":
			m.Method, err = d.Str()
		case "params":
			val, err := d.Raw()
			if err != nil {
				return err
			}
			m.Params = json.RawMessage(slices.Clone(val))
		case "result":
			val, err := d.Raw()
			if err != nil {
				return err
			}
			m.Result = io.NopCloser(bytes.NewBuffer(slices.Clone(val)))
		case "error":
			// Use the parseError helper function
			je, err := parseError(d)
			if err != nil {
				return err
			}
			m.Error = je
		}
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Message) UnmarshalJSON(xs []byte) error {
	dec := jx.GetDecoder()
	defer jx.PutDecoder(dec)
	dec.ResetBytes(xs)
	return UnmarshalMessage(m, dec)
}

func (m Message) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := jx.NewStreamingEncoder(buf, 4096)
	err := MarshalMessage(&m, enc)
	if err != nil {
		return nil, err
	}
	err = enc.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (msg *Message) String() string {
	b, _ := msg.MarshalJSON()
	return string(b)
}

// encapsulate json rpc error into struct
type JsonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (err *JsonError) Error() string {
	if err.Message == "" {
		return "json-rpc error " + strconv.Itoa(err.Code)
	}
	return err.Message
}

func (err *JsonError) ErrorCode() int {
	return err.Code
}

func (err *JsonError) ErrorData() any {
	return err.Data
}

// isBatch returns true when the first non-whitespace characters is '['
func IsBatchMessage(raw json.RawMessage) bool {
	for _, c := range raw {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		switch c {
		case 0x20, 0x09, 0x0a, 0x0d:
			continue
		}
		return c == '['
	}
	return false
}

// parseMessage parses raw bytes as a (batch of) JSON-RPC message(s). There are no error
// checks in this function because the raw message has already been syntax-checked when it
// is called. Any non-JSON-RPC messages in the input return the zero value of
// Message.
func ParseMessage(in json.RawMessage) ([]*Message, bool) {
	return ReadMessage(jx.DecodeBytes(in))
}

// parseMessage parses raw bytes as a (batch of) JSON-RPC message(s). There are no error
// checks in this function because the raw message has already been syntax-checked when it
// is called. Any non-JSON-RPC messages in the input return the zero value of
// Message.
func ReadMessage(dec *jx.Decoder) ([]*Message, bool) {
	switch dec.Next() {
	case jx.Object:
		msg := new(Message)
		err := UnmarshalMessage(msg, dec)
		if err != nil {
			msg = &Message{}
		}
		return []*Message{msg}, false
	case jx.Array:
		// Pre-allocate with a reasonable capacity
		msgs := make([]*Message, 0, 4)
		err := dec.Arr(func(d *jx.Decoder) error {
			// Check what type of value we have
			next := d.Next()
			msg := new(Message)

			// If it's not an object, it's an invalid message
			if next != jx.Object {
				// Skip the invalid value
				if err := d.Skip(); err != nil {
					return err
				}
				// Add an empty message to represent the invalid entry
				msgs = append(msgs, msg)
				return nil
			}

			// It's an object, try to unmarshal it
			UnmarshalMessage(msg, d)
			// Always append the message, even if there was an error
			// The server will handle generating the appropriate error response
			msgs = append(msgs, msg)
			return nil
		})
		if err != nil {
			return nil, true
		}
		return msgs, true
	default:
		return []*Message{{}}, false
	}
}

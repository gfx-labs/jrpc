package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-faster/jx"
)

var Null = json.RawMessage("null")

func NewNull() json.RawMessage {
	return json.RawMessage("null")
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type Message struct {
	ID     *ID             `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Error  error           `json:"error,omitempty"`

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
				EncodeError(e, m.Error)
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

func UnmarshalMessage(m *Message, dec *jx.Decoder) error {
	err := dec.Obj(func(d *jx.Decoder, key string) (err error) {
		switch key {
		case "jsonrpc":
			value, err := d.Str()
			if err != nil {
				return err
			}
			if value != VersionString {
				return NewInvalidRequestError("Invalid Version")
			}
		case "id":
			raw, err := d.Raw()
			if err != nil {
				return err
			}
			id := &ID{}
			err = id.UnmarshalJSON(raw)
			m.ID = id
			if err != nil {
				return err
			}
		case "method":
			m.Method, err = d.Str()
		case "params":
			val, err := d.RawAppend(nil)
			if err != nil {
				return err
			}
			m.Params = json.RawMessage(val)
		case "result":
			// allocate full result :)
			val, err := d.RawAppend(nil)
			if err != nil {
				return err
			}
			m.Result = io.NopCloser(bytes.NewBuffer(val))
		case "error":
			val, err := d.Raw()
			if err != nil {
				return err
			}
			m.Error = &JsonError{}
			err = json.Unmarshal(val, m.Error)
			if err != nil {
				return err
			}
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
	msgs := []*Message{{}}
	switch dec.Next() {
	case jx.Object:
		_ = UnmarshalMessage(msgs[0], dec)
		return msgs, false
	default:
		return msgs, false
	case jx.Array:
		msgs = []*Message{}
		dec.Arr(func(d *jx.Decoder) error {
			msg := new(Message)
			//err := UnmarshalMessage(msg, d)
			raw, err := d.Raw()
			if err != nil {
				raw = []byte{}
			}
			err = json.Unmarshal(raw, msg)
			if err != nil {
				msg = nil
			}
			msgs = append(msgs, msg)
			return nil
		})
		return msgs, true
	}
}

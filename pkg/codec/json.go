package codec

import (
	"encoding/json"
	"strconv"

	"github.com/go-faster/jx"

	gojson "github.com/goccy/go-json"
)

var Null = json.RawMessage("null")

func NewNull() json.RawMessage {
	return json.RawMessage("null")
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type Message struct {
	Version Version         `json:"jsonrpc,omitempty"`
	ID      *ID             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`

	Error *JsonError `json:"error,omitempty"`
}

func (m *Message) MarshalJSON() ([]byte, error) {
	var enc jx.Encoder
	// use encoder
	enc.Obj(func(e *jx.Encoder) {
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
				xs, _ := json.Marshal(m.Error)
				e.Raw(xs)
			})
		}
		if len(m.Params) != 0 {
			e.Field("params", func(e *jx.Encoder) {
				e.Raw(m.Params)
			})
		}
		if len(m.Result) != 0 {
			e.Field("result", func(e *jx.Encoder) {
				e.Raw(m.Result)
			})
		}
	})
	// output
	return enc.Bytes(), nil
}

func MakeCall(id int, method string, params []any) *Message {
	return &Message{
		ID: NewNumberIDPtr(int64(id)),
	}
}

func (msg *Message) isNotification() bool {
	return msg.ID == nil && len(msg.Method) > 0
}

func (msg *Message) String() string {
	b, _ := json.Marshal(msg)
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
func ParseMessage(raw json.RawMessage) ([]*Message, bool) {
	if !IsBatchMessage(raw) {
		msgs := []*Message{{}}
		gojson.Unmarshal(raw, &msgs[0])
		return msgs, false
	}
	// TODO:
	// for some reason other json decoders are incompatible with our test suite
	// pretty sure its how we horle EOFs and stuff
	dec := jx.DecodeBytes(raw)
	var msgs []*Message
	dec.Arr(func(d *jx.Decoder) error {
		msg := new(Message)
		raw, err := d.Raw()
		if err != nil {
			return nil
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

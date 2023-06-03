package codec

import (
	"bytes"
	"encoding/json"
	"strconv"

	"gfx.cafe/open/jrpc/contrib/codecs/websocket/wsjson"
)

var jzon = wsjson.JZON

var Null = json.RawMessage("null")

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

func MakeCall(id int, method string, params []any) *Message {
	return &Message{
		ID: NewNumberIDPtr(int64(id)),
	}
}

func (msg *Message) isNotification() bool {
	return msg.ID == nil && len(msg.Method) > 0
}

func (msg *Message) isCall() bool {
	return msg.hasValidID() && len(msg.Method) > 0
}

func (msg *Message) isResponse() bool {
	return msg.hasValidID() && len(msg.Method) == 0 && msg.Params == nil && (msg.Result != nil || msg.Error != nil)
}

func (msg *Message) hasValidID() bool {
	return msg.ID != nil && !msg.ID.IsNull()
}

func (msg *Message) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}

func (msg *Message) ErrorResponse(err error) *Message {
	resp := ErrorMessage(err)
	if resp.ID != nil {
		resp.ID = msg.ID
	}
	return resp
}
func (msg *Message) response(result any) *Message {
	// do a funny marshaling
	enc, err := jzon.Marshal(result)
	if err != nil {
		return msg.ErrorResponse(err)
	}
	if len(enc) == 0 {
		enc = []byte("null")
	}
	return &Message{ID: msg.ID, Result: enc}
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

// error message produces json rpc message with error message
func ErrorMessage(err error) *Message {
	if err == nil {
		return nil
	}
	msg := &Message{
		ID: NewNullIDPtr(),
		Error: &JsonError{
			Code:    ErrorCodeDefault,
			Message: err.Error(),
		}}
	ec, ok := err.(Error)
	if ok {
		msg.Error.Code = ec.ErrorCode()
	}
	de, ok := err.(DataError)
	if ok {
		msg.Error.Data = de.ErrorData()
	}
	return msg
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
		jzon.Unmarshal(raw, &msgs[0])
		return msgs, false
	}
	// TODO:
	// for some reason other json decoders are incompatible with our test suite
	// pretty sure its how we handle EOFs and stuff
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.Token() // skip '['
	var msgs []*Message
	for dec.More() {
		msgs = append(msgs, new(Message))
		dec.Decode(&msgs[len(msgs)-1])
	}
	return msgs, true
}

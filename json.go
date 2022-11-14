package jrpc

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	stdjson "encoding/json"

	"gfx.cafe/open/jrpc/wsjson"
	"github.com/goccy/go-json"
)

var jzon = wsjson.JZON

const (
	defaultWriteTimeout = 10 * time.Second // used if context has no deadline
)

var null = json.RawMessage("null")

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type jsonrpcMessage struct {
	Version version         `json:"jsonrpc,omitempty"`
	ID      *ID             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`

	Error *jsonError `json:"error,omitempty"`
}

func MakeCall(id int, method string, params []any) *JsonRpcMessage {
	return &JsonRpcMessage{
		ID: NewNumberIDPtr(int64(id)),
	}
}

type JsonRpcMessage = jsonrpcMessage

func (msg *jsonrpcMessage) isNotification() bool {
	return msg.ID == nil && len(msg.Method) > 0
}
func (msg *jsonrpcMessage) isCall() bool {
	return msg.hasValidID() && len(msg.Method) > 0
}
func (msg *jsonrpcMessage) isResponse() bool {
	return msg.hasValidID() && len(msg.Method) == 0 && msg.Params == nil && (msg.Result != nil || msg.Error != nil)
}
func (msg *jsonrpcMessage) hasValidID() bool {
	return msg.ID != nil && !msg.ID.null
}
func (msg *jsonrpcMessage) isSubscribe() bool {
	return strings.HasSuffix(msg.Method, subscribeMethodSuffix)
}
func (msg *jsonrpcMessage) isUnsubscribe() bool {
	return strings.HasSuffix(msg.Method, unsubscribeMethodSuffix)
}

func (msg *jsonrpcMessage) namespace() string {
	elem := strings.SplitN(msg.Method, serviceMethodSeparator, 2)
	return elem[0]
}

func (msg *jsonrpcMessage) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}

func (msg *jsonrpcMessage) errorResponse(err error) *jsonrpcMessage {
	resp := errorMessage(err)
	if resp.ID != nil {
		resp.ID = msg.ID
	}
	return resp
}
func (msg *jsonrpcMessage) response(result any) *jsonrpcMessage {
	// do a funny marshaling
	enc, err := jzon.Marshal(result)
	if err != nil {
		return msg.errorResponse(err)
	}
	if len(enc) == 0 {
		enc = []byte("null")
	}
	return &jsonrpcMessage{ID: msg.ID, Result: enc}
}

// encapsulate json rpc error into struct
type jsonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type JsonError = jsonError

func (err *jsonError) Error() string {
	if err.Message == "" {
		return "json-rpc error " + strconv.Itoa(err.Code)
	}
	return err.Message
}

func (err *jsonError) ErrorCode() int {
	return err.Code
}

func (err *jsonError) ErrorData() any {
	return err.Data
}

// error message produces json rpc message with error message
func errorMessage(err error) *jsonrpcMessage {
	msg := &jsonrpcMessage{
		ID: NewNullIDPtr(),
		Error: &jsonError{
			Code:    defaultErrorCode,
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
func isBatch(raw json.RawMessage) bool {
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
// jsonrpcMessage.
func parseMessage(raw json.RawMessage) ([]*jsonrpcMessage, bool) {
	if !isBatch(raw) {
		msgs := []*jsonrpcMessage{{}}
		json.Unmarshal(raw, &msgs[0])
		return msgs, false
	}
	// TODO:
	// for some reason other json decoders are incompatible with our test suite
	// pretty sure its how we handle EOFs and stuff
	dec := stdjson.NewDecoder(bytes.NewReader(raw))
	dec.Token() // skip '['
	var msgs []*jsonrpcMessage
	for dec.More() {
		msgs = append(msgs, new(jsonrpcMessage))
		dec.Decode(&msgs[len(msgs)-1])
	}
	return msgs, true
}

// parsePositionalArguments tries to parse the given args to an array of values with the
// given types. It returns the parsed values or an error when the args could not be
// parsed. Missing optional arguments are returned as reflect.Zero values.
func parsePositionalArguments(rawArgs json.RawMessage, types []reflect.Type) ([]reflect.Value, error) {
	var args []reflect.Value
	switch {
	case len(rawArgs) == 0:
	case rawArgs[0] == '[':
		// Read argument array.
		var err error
		if args, err = parseArgumentArray(rawArgs, types); err != nil {
			return nil, err
		}
	case string(rawArgs) == "null":
		return nil, nil
	default:
		return nil, errors.New("non-array args")
	}
	// Set any missing args to nil.
	for i := len(args); i < len(types); i++ {
		if types[i].Kind() != reflect.Ptr {
			return nil, fmt.Errorf("missing value for required argument %d", i)
		}
		args = append(args, reflect.Zero(types[i]))
	}
	return args, nil
}

func parseArgumentArray(p json.RawMessage, types []reflect.Type) ([]reflect.Value, error) {
	dec := jzon.BorrowIterator(p)
	defer jzon.ReturnIterator(dec)
	args := make([]reflect.Value, 0, len(types))
	for i := 0; dec.ReadArray(); i++ {
		if i >= len(types) {
			return args, fmt.Errorf("too many arguments, want at most %d", len(types))
		}
		argval := reflect.New(types[i])
		dec.ReadVal(argval.Interface())
		if err := dec.Error; err != nil {
			return args, fmt.Errorf("invalid argument %d: %v", i, err)
		}
		if argval.IsNil() && types[i].Kind() != reflect.Ptr {
			return args, fmt.Errorf("missing value for required argument %d", i)
		}
		args = append(args, argval.Elem())
	}
	return args, nil
}

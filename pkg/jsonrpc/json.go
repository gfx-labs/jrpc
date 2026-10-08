package jsonrpc

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/gfx-labs/jrpc/pkg/jjson"
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

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// MarshalMessage writes the JSON encoding of m to w.
func MarshalMessage(m *Message, w io.Writer) error {
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()
	if err := appendMessage(buf, m); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// AppendMessage appends the JSON encoding of m to dst.
func AppendMessage(dst []byte, m *Message) ([]byte, error) {
	buf := bytes.NewBuffer(dst)
	err := appendMessage(buf, m)
	return buf.Bytes(), err
}

func appendMessage(buf *bytes.Buffer, m *Message) error {
	buf.WriteString(`{"jsonrpc":"2.0"`)
	if m.ID != nil {
		buf.WriteString(`,"id":`)
		buf.Write(m.ID.RawMessage())
	}
	if m.Method != "" {
		buf.WriteString(`,"method":`)
		appendQuote(buf, m.Method)
	}
	if m.Error != nil {
		buf.WriteString(`,"error":`)
		buf.Write(MarshalError(m.Error))
		buf.WriteByte('}')
		return nil
	}
	if len(m.Params) != 0 {
		buf.WriteString(`,"params":`)
		buf.Write(m.Params)
	}
	if m.Result != nil {
		buf.WriteString(`,"result":`)
		n, err := buf.ReadFrom(m.Result)
		if err != nil {
			return err
		}
		if n == 0 {
			buf.WriteString(NullString)
		}
	}
	for k, v := range m.ExtraFields {
		buf.WriteByte(',')
		appendQuote(buf, k)
		buf.WriteByte(':')
		buf.Write(v)
	}
	buf.WriteByte('}')
	return nil
}

// appendQuote writes s as a JSON string. Invalid UTF-8 is replaced with U+FFFD.
func appendQuote(buf *bytes.Buffer, s string) {
	buf.Grow(len(s) + 2)
	b := buf.AvailableBuffer()
	b, _ = jsontext.AppendQuote(b, s)
	buf.Write(b)
}

// unquote returns the contents of the raw JSON string v.
func unquote(v jsontext.Value) (string, error) {
	if bytes.IndexByte(v, '\\') < 0 {
		return string(v[1 : len(v)-1]), nil
	}
	b, err := jsontext.AppendUnquote(nil, v)
	return string(b), err
}

// rawEqual reports whether the raw JSON string v equals lit, where quoted is lit wrapped in quotes.
func rawEqual(v jsontext.Value, quoted string) bool {
	if string(v) == quoted {
		return true
	}
	if bytes.IndexByte(v, '\\') < 0 {
		return false
	}
	s, err := unquote(v)
	return err == nil && s == quoted[1:len(quoted)-1]
}

// readID reads an ID value. A non-nil semErr means the value was valid JSON but not a valid ID.
func readID(d *jsontext.Decoder) (id *ID, semErr, synErr error) {
	v, err := d.ReadValue()
	if err != nil {
		return nil, nil, err
	}
	switch v.Kind() {
	case jsontext.KindNull:
		return NewNullIDPtr(), nil, nil
	case jsontext.KindNumber:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid id: %w", err), nil
		}
		return NewNumberIDPtr(n), nil, nil
	case jsontext.KindString:
		if len(v) == 2 {
			return nil, nil, nil
		}
		o := ID(slices.Clone(v))
		return &o, nil, nil
	default:
		return nil, fmt.Errorf("invalid id type"), nil
	}
}

// ParseIDBytes parses an ID from raw JSON bytes.
func ParseIDBytes(data []byte) (*ID, error) {
	d := jjson.GetDecoder(data)
	defer jjson.PutDecoder(d)
	id, semErr, err := readID(&d.Decoder)
	if err != nil {
		return nil, err
	}
	return id, semErr
}

// readError reads a JSON-RPC error object.
func readError(d *jsontext.Decoder) (je *JsonError, semErr, synErr error) {
	if d.PeekKind() != jsontext.KindBeginObject {
		if err := d.SkipValue(); err != nil {
			return nil, nil, err
		}
		return nil, fmt.Errorf("error must be an object"), nil
	}
	if _, err := d.ReadToken(); err != nil {
		return nil, nil, err
	}
	je = &JsonError{}
	for {
		switch d.PeekKind() {
		case jsontext.KindEndObject:
			if _, err := d.ReadToken(); err != nil {
				return nil, nil, err
			}
			return je, semErr, nil
		case jsontext.KindInvalid:
			_, err := d.ReadToken()
			return nil, nil, err
		}
		name, err := d.ReadValue()
		if err != nil {
			return nil, nil, err
		}
		// name is only valid until the next read, so match it first
		field := 0
		switch {
		case rawEqual(name, `"code"`):
			field = 1
		case rawEqual(name, `"message"`):
			field = 2
		case rawEqual(name, `"data"`):
			field = 3
		}
		v, err := d.ReadValue()
		if err != nil {
			return nil, nil, err
		}
		if semErr != nil {
			continue
		}
		switch field {
		case 1:
			code, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				semErr = fmt.Errorf("invalid error code: %w", err)
				continue
			}
			je.Code = int(code)
		case 2:
			if v.Kind() != jsontext.KindString {
				semErr = fmt.Errorf("invalid error message")
				continue
			}
			je.Message, err = unquote(v)
			if err != nil {
				semErr = err
			}
		case 3:
			if err := jjson.Unmarshal(v, &je.Data); err != nil {
				semErr = err
			}
		}
	}
}

// ParseErrorBytes parses a JSON-RPC error from raw JSON bytes.
func ParseErrorBytes(data []byte) (*JsonError, error) {
	d := jjson.GetDecoder(data)
	defer jjson.PutDecoder(d)
	je, semErr, err := readError(&d.Decoder)
	if err != nil {
		return nil, err
	}
	if semErr != nil {
		return nil, semErr
	}
	return je, nil
}

// readMessage reads a JSON object into m. The whole object is always consumed unless synErr is set.
// semErr reports the first JSON-RPC level problem, such as a wrong version or invalid id.
func readMessage(m *Message, d *jsontext.Decoder) (semErr, synErr error) {
	if _, err := d.ReadToken(); err != nil {
		return nil, err
	}
	for {
		switch d.PeekKind() {
		case jsontext.KindEndObject:
			_, err := d.ReadToken()
			return semErr, err
		case jsontext.KindInvalid:
			_, err := d.ReadToken()
			return nil, err
		}
		name, err := d.ReadValue()
		if err != nil {
			return nil, err
		}
		switch {
		case rawEqual(name, `"jsonrpc"`):
			v, err := d.ReadValue()
			if err != nil {
				return nil, err
			}
			if !rawEqual(v, `"2.0"`) && semErr == nil {
				semErr = NewInvalidRequestError("Invalid Version")
			}
		case rawEqual(name, `"id"`):
			id, idErr, err := readID(d)
			if err != nil {
				return nil, err
			}
			if idErr != nil {
				if semErr == nil {
					semErr = idErr
				}
				continue
			}
			m.ID = id
		case rawEqual(name, `"method"`):
			v, err := d.ReadValue()
			if err != nil {
				return nil, err
			}
			if v.Kind() != jsontext.KindString {
				if semErr == nil {
					semErr = NewInvalidRequestError("invalid method")
				}
				continue
			}
			m.Method, err = unquote(v)
			if err != nil && semErr == nil {
				semErr = err
			}
		case rawEqual(name, `"params"`):
			v, err := d.ReadValue()
			if err != nil {
				return nil, err
			}
			m.Params = json.RawMessage(slices.Clone(v))
		case rawEqual(name, `"result"`):
			v, err := d.ReadValue()
			if err != nil {
				return nil, err
			}
			m.Result = io.NopCloser(bytes.NewReader(slices.Clone(v)))
		case rawEqual(name, `"error"`):
			je, jeErr, err := readError(d)
			if err != nil {
				return nil, err
			}
			if jeErr != nil {
				if semErr == nil {
					semErr = jeErr
				}
				continue
			}
			m.Error = je
		default:
			key, err := unquote(name)
			if err != nil {
				return nil, err
			}
			v, err := d.ReadValue()
			if err != nil {
				return nil, err
			}
			if m.ExtraFields == nil {
				m.ExtraFields = make(map[string]json.RawMessage)
			}
			m.ExtraFields[key] = json.RawMessage(slices.Clone(v))
		}
	}
}

// UnmarshalMessage reads a single JSON-RPC message object from dec into m.
func UnmarshalMessage(m *Message, dec *jsontext.Decoder) error {
	if k := dec.PeekKind(); k != jsontext.KindBeginObject {
		if k == jsontext.KindInvalid {
			_, err := dec.ReadToken()
			return err
		}
		if err := dec.SkipValue(); err != nil {
			return err
		}
		return NewInvalidRequestError("message must be an object")
	}
	semErr, err := readMessage(m, dec)
	if err != nil {
		return err
	}
	return semErr
}

// UnmarshalJSONFrom implements json/v2 UnmarshalerFrom.
func (m *Message) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return UnmarshalMessage(m, dec)
}

func (m *Message) UnmarshalJSON(xs []byte) error {
	d := jjson.GetDecoder(xs)
	defer jjson.PutDecoder(d)
	return UnmarshalMessage(m, &d.Decoder)
}

func (m Message) MarshalJSON() ([]byte, error) {
	return AppendMessage(nil, &m)
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

// ParseMessage parses raw bytes as a (batch of) JSON-RPC message(s).
// It returns an error only if the input is not valid JSON. Elements that are valid JSON but
// not valid JSON-RPC messages are returned as partially filled or zero value Messages.
func ParseMessage(in json.RawMessage) ([]*Message, bool, error) {
	d := jjson.GetDecoder(in)
	defer jjson.PutDecoder(d)
	msgs, batch, err := ReadMessage(&d.Decoder)
	if err != nil {
		return nil, false, err
	}
	// reject trailing data after the top level value
	if _, err := d.ReadToken(); err != io.EOF {
		if err == nil {
			err = errors.New("invalid data after top-level value")
		}
		return nil, false, err
	}
	return msgs, batch, nil
}

// ReadMessage reads one top level JSON value from dec as a (batch of) JSON-RPC message(s).
// It returns an error only if the input is not valid JSON.
func ReadMessage(dec *jsontext.Decoder) ([]*Message, bool, error) {
	switch dec.PeekKind() {
	case jsontext.KindBeginObject:
		msg := new(Message)
		semErr, err := readMessage(msg, dec)
		if err != nil {
			return nil, false, err
		}
		if semErr != nil {
			msg = &Message{}
		}
		return []*Message{msg}, false, nil
	case jsontext.KindBeginArray:
		if _, err := dec.ReadToken(); err != nil {
			return nil, true, err
		}
		msgs := make([]*Message, 0, 4)
		for {
			msg := new(Message)
			switch dec.PeekKind() {
			case jsontext.KindEndArray:
				if _, err := dec.ReadToken(); err != nil {
					return nil, true, err
				}
				return msgs, true, nil
			case jsontext.KindInvalid:
				_, err := dec.ReadToken()
				return nil, true, err
			case jsontext.KindBeginObject:
				// keep partially parsed messages; the server generates the error response
				if _, err := readMessage(msg, dec); err != nil {
					return nil, true, err
				}
			default:
				if err := dec.SkipValue(); err != nil {
					return nil, true, err
				}
			}
			msgs = append(msgs, msg)
		}
	case jsontext.KindInvalid:
		_, err := dec.ReadToken()
		return nil, false, err
	default:
		if err := dec.SkipValue(); err != nil {
			return nil, false, err
		}
		return []*Message{{}}, false, nil
	}
}

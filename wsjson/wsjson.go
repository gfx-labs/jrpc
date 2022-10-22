package wsjson

import (
	"context"
	"fmt"

	"gfx.cafe/util/go/bufpool"
	jsoniter "github.com/json-iterator/go"
	"nhooyr.io/websocket"
)

var jzon = jsoniter.Config{
	IndentionStep:                 0,
	MarshalFloatWith6Digits:       false,
	EscapeHTML:                    true,
	SortMapKeys:                   true,
	UseNumber:                     false,
	DisallowUnknownFields:         false,
	TagKey:                        "",
	OnlyTaggedField:               false,
	ValidateJsonRawMessage:        false,
	ObjectFieldMustBeSimpleString: false,
	CaseSensitive:                 false,
}.Froze()

var JZON = jzon
var JSON = jsoniter.Config{
	IndentionStep:                 0,
	MarshalFloatWith6Digits:       false,
	EscapeHTML:                    true,
	SortMapKeys:                   true,
	UseNumber:                     false,
	DisallowUnknownFields:         false,
	TagKey:                        "",
	OnlyTaggedField:               false,
	ValidateJsonRawMessage:        false,
	ObjectFieldMustBeSimpleString: false,
	CaseSensitive:                 false,
}.Froze()

// Read reads a JSON message from c into v.
// It will reuse buffers in between calls to avoid allocations.
func Read(ctx context.Context, c *websocket.Conn, v interface{}) error {
	return read(ctx, c, v)
}

func read(ctx context.Context, c *websocket.Conn, v interface{}) (err error) {
	_, r, err := c.Reader(ctx)
	if err != nil {
		return err
	}
	b := bufpool.GetStd()
	defer bufpool.PutStd(b)
	_, err = b.ReadFrom(r)
	if err != nil {
		return err
	}
	err = jzon.NewDecoder(b).Decode(v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// Write writes the JSON message v to c.
// It will reuse buffers in between calls to avoid allocations.
func Write(ctx context.Context, c *websocket.Conn, v interface{}) error {
	return write(ctx, c, v)
}

func write(ctx context.Context, c *websocket.Conn, v interface{}) (err error) {
	w, err := c.Writer(ctx, websocket.MessageText)
	if err != nil {
		return err
	}
	// json.Marshal cannot reuse buffers between calls as it has to return
	// a copy of the byte slice but Encoder does as it directly writes to w.
	st := jsoniter.NewStream(jzon, w, 1024)
	st.WriteVal(v)
	err = st.Flush()
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return w.Close()
}

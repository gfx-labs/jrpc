package jsonrpc

import (
	"encoding/json"
	"io"

	"github.com/go-faster/jx"
)

// MessageStream is a writer used to write jsonrpc message to a stream
type MessageStream struct {
	w  io.Writer
	jx *jx.Writer
}

func NewStream(w io.Writer) (*MessageStream, error) {
	enc := jx.GetWriter()
	defer jx.PutWriter(enc)
	enc.Grow(4096)
	enc.ResetWriter(w)
	enc.ObjStart()
	enc.FieldStart("jsonrpc")
	enc.Str("2.0")
	enc.Close()
	return &MessageStream{
		w:  w,
		jx: enc,
	}, nil
}

func (m *MessageStream) Field(name string, value json.RawMessage) error {
	m.jx.ResetWriter(m.w)
	m.jx.Comma()
	m.jx.FieldStart(name)
	m.jx.Raw(value)
	return m.jx.Close()
}

// Result returns a writecloser that writes to a result field
func (m *MessageStream) Result() (io.Writer, error) {
	m.jx.ResetWriter(m.w)
	m.jx.Comma()
	m.jx.FieldStart("result")
	m.jx.Close()
	return &MessageWriter{w: m.w}, nil
}

func (m *MessageStream) Close() error {
	_, err := m.w.Write([]byte("}"))
	return err
}

type MessageWriter struct {
	w io.Writer
}

func (m *MessageWriter) Write(p []byte) (n int, err error) {
	return m.w.Write(p)
}

package jsonrpc

import (
	"encoding/json"
	"io"
)

// MessageStream is a writer used to write jsonrpc message to a stream
type MessageStream struct {
	w io.Writer
}

func NewStream(w io.Writer) *MessageStream {
	return &MessageStream{
		w: w,
	}
}

func (m *MessageStream) NewMessage() (*MessageWriter, error) {
	_, err := m.w.Write([]byte(`{"jsonrpc":"2.0"`))
	if err != nil {
		return nil, err
	}
	return &MessageWriter{
		w: m.w,
	}, nil
}

type MessageWriter struct {
	w io.Writer
}

func (m *MessageWriter) Field(name string, value json.RawMessage) error {
	_, err := m.w.Write([]byte(`,"` + name + `":`))
	if err != nil {
		return err
	}
	_, err = m.w.Write(value)
	if err != nil {
		return err
	}
	return nil
}

// Result returns a writer that writes to a result field
func (m *MessageWriter) Result() (io.Writer, error) {
	_, err := m.w.Write([]byte(`,"result":`))
	if err != nil {
		return nil, err
	}
	return &ResultWriter{w: m.w}, nil
}

func (m *MessageWriter) Close() error {
	_, err := m.w.Write([]byte("}"))
	return err
}

type ResultWriter struct {
	w io.Writer
}

func (m *ResultWriter) Write(p []byte) (n int, err error) {
	return m.w.Write(p)
}

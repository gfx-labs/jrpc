package jsonrpc

import (
	"encoding/json"
	"io"

	"golang.org/x/net/context"
	"golang.org/x/sync/semaphore"
)

// MessageStream is a writer used to write jsonrpc message to a stream
type MessageStream struct {
	w  io.Writer
	mu *semaphore.Weighted
}

func NewStream(w io.Writer) *MessageStream {
	return &MessageStream{
		w:  w,
		mu: semaphore.NewWeighted(1),
	}
}

// NewMessage starts a new message and acquires the write lock.
// to free the write lock, you must call *MessageWriter.Close()
// the lock MUST be closed if and only if err != nil
func (m *MessageStream) NewMessage(ctx context.Context) (*MessageWriter, error) {
	err := m.mu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	_, err = m.w.Write([]byte(`{"jsonrpc":"2.0"`))
	if err != nil {
		m.mu.Release(1)
		return nil, err
	}
	return &MessageWriter{
		w:  m.w,
		mu: m.mu,
	}, nil
}

type MessageWriter struct {
	w  io.Writer
	mu *semaphore.Weighted
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

// close must be called when you are done writing the message.
// it releases the write lock
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

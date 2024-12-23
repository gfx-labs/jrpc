package jsonrpc

import (
	"encoding/json"
	"io"
	"sync"

	"golang.org/x/net/context"
)

type MessageStreamer interface {
	NewMessage(ctx context.Context) (*MessageWriter, error)
}
type flusher interface {
	Flush() error
}

func flushIfFlusher(w io.Writer) error {
	if val, ok := w.(flusher); ok {
		return val.Flush()
	}
	return nil
}

// MessageStream is a writer used to write jsonrpc message to a stream
type MessageStream struct {
	w  io.Writer
	mu *sync.Mutex
}

func NewStream(w io.Writer) *MessageStream {
	return &MessageStream{
		w:  w,
		mu: &sync.Mutex{},
	}
}

// sends a flush in order to send an empty payload
func (m *MessageStream) Flush(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return flushIfFlusher(m.w)
}

// ReadFrom calls io.Copy within the semaphore, then calls flush
func (m *MessageStream) ReadFrom(ctx context.Context, r io.Reader) error {
	if m.mu != nil {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	_, err := io.Copy(m.w, r)
	if err != nil {
		return err
	}
	return flushIfFlusher(m.w)
}

type MessageWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

// NewMessage starts a new message and acquires the write lock.
// to free the write lock, you must call *MessageWriter.Close()
// the lock MUST be closed if and only if err == nil
func (m *MessageStream) NewMessage(ctx context.Context) (*MessageWriter, error) {
	if m.mu != nil {
		m.mu.Lock()
	}
	_, err := m.w.Write([]byte(`{"jsonrpc":"2.0"`))
	if err != nil {
		if m.mu != nil {
			m.mu.Unlock()
		}
		return nil, err
	}
	return &MessageWriter{
		w:  m.w,
		mu: m.mu,
	}, nil
}

// close must be called when you are done writing the message.
// it releases the write lock
func (m *MessageWriter) Close() error {
	if m.mu != nil {
		defer m.mu.Unlock()
	}
	_, err := m.w.Write([]byte("}"))
	if err != nil {
		return err
	}
	return flushIfFlusher(m.w)
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
func (m *MessageWriter) Result() (io.WriteCloser, error) {
	_, err := m.w.Write([]byte(`,"result":`))
	if err != nil {
		return nil, err
	}
	return &ResultWriter{w: m.w}, nil
}

// Params returns a writer that writes to a params field
func (m *MessageWriter) Params() (io.WriteCloser, error) {
	_, err := m.w.Write([]byte(`,"params":`))
	if err != nil {
		return nil, err
	}
	return &ResultWriter{w: m.w}, nil
}

type BatchWriter struct {
	w          io.Writer
	mu         *sync.Mutex
	ms         *MessageStream
	isNotFirst bool
}

type writer struct {
	w io.Writer
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

// Start writing a batch to the stream. this function acquires the lock
// caller MUST call Close() on the BatchWriter iff err == nil
func (m *MessageStream) NewBatch(ctx context.Context) (*BatchWriter, error) {
	if m.mu != nil {
		m.mu.Lock()
	}
	_, err := m.w.Write([]byte("["))
	if err != nil {
		if m.mu != nil {
			m.mu.Unlock()
		}
		return nil, err
	}
	return &BatchWriter{
		w: m.w,
		ms: &MessageStream{
			// we wrap the writer here with a noflush writer so we can reuse the messagestream
			// when the messagestream creates its subwrites, they won't pass the interface check for Flush
			// so they wont flush when they close.
			w: &writer{m.w},
			// the mutex is nil because we are going to use the mutex from the message stream.
			mu: nil,
		},
		mu: m.mu,
	}, nil
}

// Writes the next element in the batch. Note that the messagewriter is not thread safe
func (m *BatchWriter) NewMessage(ctx context.Context) (*MessageWriter, error) {
	if m.isNotFirst == false {
		m.isNotFirst = true
	} else {
		// write comma if not the first element
		_, err := m.w.Write([]byte(","))
		if err != nil {
			return nil, err
		}
	}
	return m.ms.NewMessage(ctx)
}

// close must be called when you are done writing the batch.
// it releases the write lock
func (m *BatchWriter) Close() error {
	if m.mu != nil {
		defer m.mu.Unlock()
	}
	_, err := m.w.Write([]byte("]"))
	if err != nil {
		return err
	}
	return flushIfFlusher(m.w)
}

type ResultWriter struct {
	w     io.Writer
	wrote bool
}

func (m *ResultWriter) Write(p []byte) (n int, err error) {
	m.wrote = true
	return m.w.Write(p)
}

func (m *ResultWriter) Close() error {
	if m.wrote == false {
		_, err := m.w.Write(Null)
		return err
	}
	return nil
}

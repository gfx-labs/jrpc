package jrpc

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/goccy/go-json"
)

// DeadlineConn is a subset of the methods of net.Conn which are sufficient for creating a jsonCodec
type DeadlineConn interface {
	io.ReadWriteCloser
	SetWriteDeadline(time.Time) error
}

// jsonCodec reads and writes JSON-RPC messages to the underlying connection. It also has
// support for parsing arguments and serializing (result) objects.
type jsonCodec struct {
	remote    string
	closer    sync.Once // close closed channel once
	closeFunc func() error
	closeCh   chan any          // closed on Close
	decode    func(v any) error // decoder to allow multiple transports
	encMu     sync.Mutex        // guards the encoder
	encode    func(v any) error // encoder to allow multiple transports
	conn      DeadlineConn
}

// NewFuncCodec creates a codec which uses the given functions to read and write. If conn
// implements ConnRemoteAddr, log messages will use it to include the remote address of
// the connection.
func NewFuncCodec(
	conn DeadlineConn,
	encode, decode func(v any) error,
	closeFunc func() error,
) ServerCodec {
	codec := &jsonCodec{
		closeFunc: closeFunc,
		closeCh:   make(chan any),
		encode:    encode,
		decode:    decode,
		conn:      conn,
	}
	if ra, ok := conn.(interface{ RemoteAddr() string }); ok {
		codec.remote = ra.RemoteAddr()
	}
	return codec
}

// NewCodec creates a codec on the given connection. If conn implements ConnRemoteAddr, log
// messages will use it to include the remote address of the connection.
func NewCodec(conn DeadlineConn) ServerCodec {
	encr := func(v any) error {
		enc := jzon.BorrowStream(conn)
		defer jzon.ReturnStream(enc)
		enc.WriteVal(v)
		enc.WriteRaw("\n")
		enc.Flush()
		if enc.Error != nil {
			return enc.Error
		}
		return nil
	}
	// TODO:
	// for some reason other json decoders are incompatible with our test suite
	// pretty sure its how we handle EOFs and stuff
	dec := json.NewDecoder(conn)
	dec.UseNumber()
	return NewFuncCodec(conn, encr, dec.Decode, func() error {
		return nil
	})
}

func (c *jsonCodec) PeerInfo() PeerInfo {
	// This returns "ipc" because all other built-in transports have a separate codec type.
	return PeerInfo{Transport: "ipc", RemoteAddr: c.remote}
}

func (c *jsonCodec) RemoteAddr() string {
	return c.remote
}

func (c *jsonCodec) ReadBatch() (messages []*jsonrpcMessage, batch bool, err error) {
	// Decode the next JSON object in the input stream.
	// This verifies basic syntax, etc.
	var rawmsg json.RawMessage
	if err := c.decode(&rawmsg); err != nil {
		return nil, false, err
	}
	messages, batch = parseMessage(rawmsg)
	for i, msg := range messages {
		if msg == nil {
			// Message is JSON 'null'. Replace with zero value so it
			// will be treated like any other invalid message.
			messages[i] = new(jsonrpcMessage)
		}
	}
	return messages, batch, nil
}

func (c *jsonCodec) WriteJSON(ctx context.Context, v any) error {
	c.encMu.Lock()
	defer c.encMu.Unlock()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(defaultWriteTimeout)
	}
	c.conn.SetWriteDeadline(deadline)
	return c.encode(v)
}

func (c *jsonCodec) Close() error {
	c.closer.Do(func() {
		close(c.closeCh)
		if c.closeFunc != nil {
			c.closeFunc()
		}
		c.conn.Close()
	})
	return nil
}

// Closed returns a channel which will be closed when Close is called
func (c *jsonCodec) Closed() <-chan any {
	return c.closeCh
}

package inproc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"gfx.cafe/open/jrpc/codec"
)

type Codec struct {
	done chan any

	rd   io.Reader
	wr   io.Writer
	msgs chan json.RawMessage
}

func NewCodec() *Codec {
	rd, wr := io.Pipe()
	return &Codec{
		done: make(chan interface{}),
		rd:   bufio.NewReader(rd),
		wr:   wr,
		msgs: make(chan json.RawMessage, 8),
	}
}

// gets the peer info
func (c *Codec) PeerInfo() codec.PeerInfo {
	return codec.PeerInfo{
		Transport:  "ipc",
		RemoteAddr: "",
		HTTP:       codec.HttpInfo{},
	}
}

// json.RawMessage can be an array of requests. if it is, then it is a batch request
func (c *Codec) ReadBatch(ctx context.Context) (msgs json.RawMessage, err error) {
	select {
	case ans := <-c.msgs:
		return ans, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// closes the connection
func (c *Codec) Close() error {
	close(c.done)
	return nil
}

func (c *Codec) Write(p []byte) (n int, err error) {
	return c.wr.Write(p)
}

// Closed returns a channel which is closed when the connection is closed.
func (c *Codec) Closed() <-chan any {
	return c.done
}

// RemoteAddr returns the peer address of the connection.
func (c *Codec) RemoteAddr() string {
	return ""
}

// DialInProc attaches an in-process connection to the given RPC server.
//func DialInProc(handler *Server) *Client {
//	initctx := context.Background()
//	c, _ := newClient(initctx, func(context.Context) (ServerCodec, error) {
//		p1, p2 := net.Pipe()
//		go handler.ServeCodec(NewCodec(p1))
//		return NewCodec(p2), nil
//	})
//	return c
//}

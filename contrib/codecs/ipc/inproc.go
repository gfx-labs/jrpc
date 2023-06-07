package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"gfx.cafe/open/jrpc/pkg/codec"
)

type Codec struct {
	ctx context.Context
	cn  func()

	rd   io.Reader
	wr   *bufio.Writer
	msgs chan json.RawMessage
}

func NewCodec() *Codec {
	rd, wr := io.Pipe()
	ctx, cn := context.WithCancel(context.TODO())
	return &Codec{
		ctx:  ctx,
		cn:   cn,
		rd:   bufio.NewReader(rd),
		wr:   bufio.NewWriter(wr),
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
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

// closes the connection
func (c *Codec) Close() error {
	c.cn()
	return nil
}

func (c *Codec) Write(p []byte) (n int, err error) {
	return c.wr.Write(p)
}

func (c *Codec) Flush() (err error) {
	return c.wr.Flush()
}

// Closed returns a channel which is closed when the connection is closed.
func (c *Codec) Closed() <-chan struct{} {
	return c.ctx.Done()
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

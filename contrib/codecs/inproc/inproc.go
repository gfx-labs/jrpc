package inproc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/serverutil"
)

type Codec struct {
	ctx context.Context
	cn  func()

	rd     io.Reader
	wrLock sync.Mutex
	wr     *bufio.Writer
	msgs   chan *serverutil.Bundle
}

func NewCodec() *Codec {
	rd, wr := io.Pipe()
	ctx, cn := context.WithCancel(context.TODO())
	return &Codec{
		ctx:  ctx,
		cn:   cn,
		rd:   bufio.NewReader(rd),
		wr:   bufio.NewWriter(wr),
		msgs: make(chan *serverutil.Bundle, 8),
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

func (c *Codec) ReadBatch(ctx context.Context) ([]*codec.Message, bool, error) {
	select {
	case ans := <-c.msgs:
		return ans.Messages, ans.Batch, nil
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	}
}

// closes the connection
func (c *Codec) Close() error {
	c.cn()
	return nil
}

func (c *Codec) Send(ctx context.Context, msg json.RawMessage) (err error) {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	_, err = c.wr.Write(msg)
	if err != nil {
		return err
	}
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
// func DialInProc(handler *Server) *Client {
//	initctx := context.Background()
//	c, _ := newClient(initctx, func(context.Context) (ServerCodec, error) {
//		p1, p2 := net.Pipe()
//		go handler.ServeCodec(NewCodec(p1))
//		return NewCodec(p2), nil
//	})
//	return c
// }

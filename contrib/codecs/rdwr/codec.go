package rdwr

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/gfx-labs/jrpc/pkg/jjson"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
	"github.com/gfx-labs/jrpc/pkg/serverutil"
)

type Codec struct {
	ctx context.Context
	cn  func()

	rd io.Reader
	wr *bufio.Writer

	decLock sync.Mutex
}

func NewCodec(rd io.Reader, wr io.Writer) *Codec {
	ctx, cn := context.WithCancel(context.TODO())
	c := &Codec{
		ctx: ctx,
		cn:  cn,
		rd:  bufio.NewReader(rd),
		wr:  bufio.NewWriter(wr),
	}
	return c
}

// gets the peer info
func (c *Codec) PeerInfo() jsonrpc.PeerInfo {
	return jsonrpc.PeerInfo{
		Transport:  "ipc",
		RemoteAddr: "",
		HTTP:       nil,
	}
}

func (c *Codec) decodeSingleMessage(ctx context.Context) (*serverutil.SimpleBundle, error) {
	c.decLock.Lock()
	defer c.decLock.Unlock()
	decBuf := make(json.RawMessage, 0)
	err := jjson.Decode(c.rd, &decBuf)
	if err != nil {
		return nil, err
	}
	return serverutil.ParseBundle(decBuf), nil
}

func (c *Codec) ReadBatch(ctx context.Context) (jsonrpc.Bundle, error) {
	ans, err := c.decodeSingleMessage(ctx)
	if err != nil {
		return nil, err
	}
	return ans, nil
}

// closes the connection
func (c *Codec) Close() error {
	c.cn()
	return nil
}

func (c *Codec) Write(p []byte) (n int, err error) {
	n, err = c.wr.Write(p)
	return n, err
}

func (c *Codec) Flush() error {
	if _, err := c.wr.Write([]byte("\n")); err != nil {
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

// Dialrdwr attaches an in-process connection to the given RPC server.
// func Dialrdwr(handler *Server) *Client {
//	initctx := context.Background()
//	c, _ := newClient(initctx, func(context.Context) (ServerCodec, error) {
//		p1, p2 := net.Pipe()
//		go handler.ServeCodec(NewCodec(p1))
//		return NewCodec(p2), nil
//	})
//	return c
// }

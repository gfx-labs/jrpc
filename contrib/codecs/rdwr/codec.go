package rdwr

import (
	"bufio"
	"context"
	"io"
	"sync"

	"github.com/go-faster/jx"
	"github.com/goccy/go-json"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/serverutil"
)

type Codec struct {
	ctx context.Context
	cn  func()

	rd     io.Reader
	wrLock sync.Mutex
	wr     *bufio.Writer
	jx     *jx.Encoder

	dec     *json.Decoder
	decBuf  json.RawMessage
	decLock sync.Mutex
}

func NewCodec(rd io.Reader, wr io.Writer) *Codec {
	ctx, cn := context.WithCancel(context.TODO())
	c := &Codec{
		ctx: ctx,
		cn:  cn,
		rd:  bufio.NewReader(rd),
		wr:  bufio.NewWriter(wr),
		dec: json.NewDecoder(rd),
	}
	c.jx = jx.NewStreamingEncoder(wr, 4096)
	return c
}

// gets the peer info
func (c *Codec) PeerInfo() codec.PeerInfo {
	return codec.PeerInfo{
		Transport:  "ipc",
		RemoteAddr: "",
		HTTP:       codec.HttpInfo{},
	}
}

func (c *Codec) decodeSingleMessage(ctx context.Context) (*serverutil.Bundle, error) {
	c.decLock.Lock()
	defer c.decLock.Unlock()
	c.decBuf = c.decBuf[:0]
	err := c.dec.DecodeContext(ctx, &c.decBuf)
	if err != nil {
		return nil, err
	}
	return serverutil.ParseBundle(c.decBuf), nil
}

func (c *Codec) ReadBatch(ctx context.Context) ([]*codec.Message, bool, error) {
	ans, err := c.decodeSingleMessage(ctx)
	if err != nil {
		return nil, false, err
	}
	return ans.Messages, ans.Batch, nil
}

// closes the connection
func (c *Codec) Close() error {
	c.cn()
	return nil
}

func (c *Codec) Send(fn func(e *jx.Encoder) error) error {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	defer c.jx.ResetWriter(c.wr)
	if err := fn(c.jx); err != nil {
		return err
	}
	if err := c.jx.Close(); err != nil {
		return err
	}
	if _, err := c.wr.Write([]byte("\n")); err != nil {
		return err
	}
	if err := c.wr.Flush(); err != nil {
		return err
	}
	return nil
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

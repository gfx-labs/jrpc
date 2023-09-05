package rdwr

import (
	"bufio"
	"context"
	"io"
	"sync"

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
	msgs   chan *serverutil.Bundle
}

func NewCodec(rd io.Reader, wr io.Writer, onError func(error)) *Codec {
	ctx, cn := context.WithCancel(context.TODO())
	c := &Codec{
		ctx:  ctx,
		cn:   cn,
		rd:   bufio.NewReader(rd),
		wr:   bufio.NewWriter(wr),
		msgs: make(chan *serverutil.Bundle, 8),
	}
	go func() {
		err := c.listen()
		if err != nil && onError != nil {
			onError(err)
		}
	}()
	return c
}

func (c *Codec) listen() error {
	var msg json.RawMessage
	dec := json.NewDecoder(c.rd)
	for {
		// reading a message
		err := dec.Decode(&msg)
		if err != nil {
			c.cn()
			return err
		}
		c.msgs <- serverutil.ParseBundle(msg)
		msg = msg[:0]
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

func (c *Codec) Write(p []byte) (n int, err error) {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	return c.wr.Write(p)
}

func (c *Codec) Flush() (err error) {
	c.wrLock.Lock()
	defer c.wrLock.Unlock()
	if c.wr.Buffered() > 0 {
		err = c.wr.WriteByte('\n')
		if err != nil {
			return err
		}
		err = c.wr.Flush()
		if err != nil {
			return err
		}
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

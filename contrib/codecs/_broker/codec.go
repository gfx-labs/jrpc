package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/serverutil"
	"github.com/gogo/protobuf/io"
)

var _ jsonrpc.ReaderWriter = (*Codec)(nil)

type Codec struct {
	ctx context.Context
	cn  func()

	wr      bytes.Buffer
	replier Replier
	ansCh   chan *serverutil.Bundle
	closed  atomic.Bool
	closeCh chan struct{}

	i jsonrpc.PeerInfo
}

type httpError struct {
	code int
	err  error
}

func NewCodec(req json.RawMessage, replier Replier) *Codec {
	c := &Codec{
		replier: replier,
		ansCh:   make(chan *serverutil.Bundle, 1),
		closeCh: make(chan struct{}),
	}
	c.ctx, c.cn = context.WithCancel(context.Background())
	bundle := serverutil.ParseBundle(req)
	c.ansCh <- bundle
	return c
}

// gets the peer info
func (c *Codec) PeerInfo() jsonrpc.PeerInfo {
	return c.i
}

func (c *Codec) ReadBatch(ctx context.Context) ([]*jsonrpc.Message, bool, error) {
	select {
	case ans := <-c.ansCh:
		return ans.Messages, ans.Batch, nil
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	}
}

// closes the connection
func (c *Codec) Close() error {
	if c.closed.CompareAndSwap(false, true) {
		close(c.closeCh)
	}
	c.cn()
	return nil
}

func (c *Codec) Send(fn func(io.Writer) error) error {
	return c.replier.Send(fn)
}

func (c *Codec) Flush() error {
	return c.replier.Send(fn)
}

// Closed returns a channel which is closed when the connection is closed.
func (c *Codec) Closed() <-chan struct{} {
	return c.closeCh
}

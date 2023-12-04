package server

import (
	"context"
	"sync"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

// 128mb... should be more than enough for any batch.
// you shouldn't be batching more than this. really, you shouldn't be using batching at all.
// TODO: make this configurable
const maxBatchSizeBytes = 1024 * 1024 * 1024 * 128

var _ jsonrpc.ResponseWriter = (*streamingRespWriter)(nil)

// streamingRespWriter is NOT thread safe
type streamingRespWriter struct {
	cr  *callResponder
	msg *jsonrpc.Message
	ctx context.Context

	err error

	sendCalled bool

	mu sync.Mutex
}

func (c *streamingRespWriter) Send(v any, e error) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.msg.ID == nil {
		return jsonrpc.ErrCantSendNotification
	}
	if c.sendCalled {
		return jsonrpc.ErrSendAlreadyCalled
	}
	c.sendCalled = true
	ce := &callEnv{
		err: c.err,
		id:  c.msg.ID,
	}
	// only override error if not already set
	if ce.err == nil {
		ce.err = e
	}
	// only set value if value is not nil
	if v != nil {
		ce.v = v
	}
	msg, err := c.cr.stream.NewMessage(c.ctx)
	if err != nil {
		return err
	}
	defer msg.Close()
	if err = send(ce, msg); err != nil {
		return err
	}
	return nil
}

func (c *streamingRespWriter) Notify(method string, v any) error {
	msg, err := c.cr.stream.NewMessage(c.ctx)
	if err != nil {
		return err
	}
	defer msg.Close()
	err = notify(&notifyEnv{
		method: method,
		dat:    v,
	}, msg)
	if err != nil {
		return err
	}
	return nil
}

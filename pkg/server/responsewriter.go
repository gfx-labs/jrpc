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
	err = c.cr.mu.Acquire(c.ctx, 1)
	if err != nil {
		return err
	}
	defer c.cr.mu.Release(1)
	if c.err != nil {
		e = c.err
	}
	if err = c.cr.send(c.ctx, ce); err != nil {
		return err
	}
	if err = c.cr.remote.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *streamingRespWriter) Notify(method string, v any) error {
	err := c.cr.mu.Acquire(c.ctx, 1)
	if err != nil {
		return err
	}
	defer c.cr.mu.Release(1)
	err = c.cr.notify(c.ctx, &notifyEnv{
		method: method,
		dat:    v,
	})
	if err != nil {
		return err
	}
	err = c.cr.remote.Flush()
	if err != nil {
		return err
	}
	return nil
}

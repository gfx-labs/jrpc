package server

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

// 128mb... should be more than enough for any batch.
// you shouldn't be batching more than this. really, you shouldn't be using batching at all.
// TODO: make this configurable
const maxBatchSizeBytes = 1024 * 1024 * 1024 * 128

var _ jsonrpc.ResponseWriter = (*streamingRespWriter)(nil)

// streamingRespWriter is NOT thread safe
type streamingRespWriter struct {
	// this should be the same context as the request
	ctx context.Context
	// the stream that Send will write to
	sendStream jsonrpc.MessageStreamer
	// the stream that Notify will write to
	notifyStream jsonrpc.MessageStreamer
	// the id to write the response with
	id *jsonrpc.ID

	// a function that is called on the first call to send
	// it's optional
	done func()

	// if set, will ensure that send will always send this error, instead of whatever send does
	err error
	// marks whether or not send was called. it may only be called once
	sendCalled bool
}

func (c *streamingRespWriter) SendStream() jsonrpc.MessageStreamer {
	if c.sendCalled {
		return c.sendStream
	}
	c.sendCalled = true
	if c.done != nil {
		c.done()
	}
	return c.sendStream
}

func (c *streamingRespWriter) NotifyStream() jsonrpc.MessageStreamer {
	return c.sendStream
}

func (c *streamingRespWriter) Send(v any, e error) (err error) {
	if c.id == nil {
		return jsonrpc.ErrCantSendNotification
	}
	if c.sendCalled {
		return jsonrpc.ErrSendAlreadyCalled
	}
	if c.done != nil {
		c.done()
	}
	c.sendCalled = true
	ce := &callEnv{
		err: c.err,
		id:  c.id,
	}
	// only override error if not already set
	if ce.err == nil {
		ce.err = e
	}
	// only set value if value is not nil
	if v != nil {
		ce.v = v
	}
	msg, err := c.sendStream.NewMessage(c.ctx)
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
	msg, err := c.notifyStream.NewMessage(c.ctx)
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

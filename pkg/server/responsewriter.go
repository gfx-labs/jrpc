package server

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

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

func (c *streamingRespWriter) SendStream(fn func(jsonrpc.MessageStreamer) error) error {
	if c.sendCalled {
		return jsonrpc.ErrSendAlreadyCalled
	}
	c.sendCalled = true
	if c.done != nil {
		defer c.done()
	}
	return fn(c.sendStream)
}

func (c *streamingRespWriter) NotifyStream(fn func(jsonrpc.MessageStreamer) error) error {
	return fn(c.notifyStream)
}

func (c *streamingRespWriter) Send(v any, e error) (err error) {
	if c.id == nil {
		return jsonrpc.ErrCantSendNotification
	}
	if c.sendCalled {
		return jsonrpc.ErrSendAlreadyCalled
	}
	if c.done != nil {
		defer c.done()
	}
	c.sendCalled = true
	sentErr := c.err
	// only override error if not already set
	if sentErr == nil {
		sentErr = e
	}
	msg, err := c.sendStream.NewMessage(c.ctx)
	if err != nil {
		return err
	}
	defer msg.Close()
	if c.id != nil {
		msg.Field("id", c.id.RawMessage())
	}
	if sentErr != nil {
		msg.Field("error", jsonrpc.MarshalError(sentErr))
		return nil
	}
	// if there is no error, we try to marshal the result
	wr, err := msg.Result()
	if err != nil {
		return err
	}
	defer wr.Close()
	return jsonrpc.EncodeObject(wr, v)
}

func (c *streamingRespWriter) Notify(method string, v any) error {
	msg, err := c.notifyStream.NewMessage(c.ctx)
	if err != nil {
		return err
	}
	defer msg.Close()
	dat := v
	err = msg.Field("method", []byte(`"`+method+`"`))
	if err != nil {
		return err
	}
	// if there is no error, we try to marshal the result
	wr, err := msg.Params()
	if err != nil {
		return err
	}
	defer wr.Close()
	return jsonrpc.EncodeObject(wr, dat)
}

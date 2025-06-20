package server

import (
	"context"
	"encoding/json"

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

	// if set, will ensure that send will always send this error, instead of whatever send does
	err error
	// marks whether or not send was called. it may only be called once
	sendCalled bool
	// marks whether or not hijack was called
	hijackCalled bool
	// extensions to add to the response
	extensions map[string]json.RawMessage
}

func (c *streamingRespWriter) Hijack() (sender jsonrpc.MessageStreamer, notify jsonrpc.MessageStreamer, err error) {
	if c.hijackCalled {
		return nil, nil, jsonrpc.ErrHijackAlreadyCalled
	}
	c.hijackCalled = true
	c.sendCalled = true
	return c.sendStream, c.notifyStream, nil
}

func (c *streamingRespWriter) Send(v any, e error) (err error) {
	if c.hijackCalled {
		return jsonrpc.ErrHijackAlreadyCalled
	}
	if c.id == nil {
		return jsonrpc.ErrCantSendNotification
	}
	if c.sendCalled {
		return jsonrpc.ErrSendAlreadyCalled
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
	// Write extensions before result/error
	for key, val := range c.extensions {
		err = msg.Field(key, val)
		if err != nil {
			return err
		}
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

	if c.hijackCalled {
		return jsonrpc.ErrHijackAlreadyCalled
	}
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

func (c *streamingRespWriter) Extension(key string, v any) error {
	if c.sendCalled {
		return jsonrpc.ErrSendAlreadyCalled
	}
	if c.hijackCalled {
		return jsonrpc.ErrHijackAlreadyCalled
	}
	if c.extensions == nil {
		c.extensions = make(map[string]json.RawMessage)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.extensions[key] = json.RawMessage(data)
	return nil
}

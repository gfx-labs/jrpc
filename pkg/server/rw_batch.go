package server

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"gfx.cafe/open/jrpc/pkg/jjson"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

// batchingRespWriter is NOT thread safe
type batchingRespWriter struct {
	cr  *callResponder
	msg *jsonrpc.Message
	ctx context.Context

	wg      *sync.WaitGroup
	payload json.RawMessage
	err     error

	sendCalled bool

	mu sync.Mutex
}

func (c *batchingRespWriter) Send(v any, e error) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.msg.ID == nil {
		return jsonrpc.ErrCantSendNotification
	}
	if c.sendCalled {
		return jsonrpc.ErrSendAlreadyCalled
	}
	c.sendCalled = true
	if c.wg != nil {
		defer c.wg.Done()
	}
	// if there is an error, and no c.err is set, and there is an e, then set c.err to e
	if c.err == nil {
		c.err = e
	}
	// batch requests are not individually streamed.
	// the reason is beacuse i couldn't think of a good way to implement it
	// ultimately they need to be buffered. there's some optimistic multiplexing you can
	// do, but that felt really complicated and not worth the time.
	if v != nil && c.err == nil {
		buf := &bytes.Buffer{}
		w := newWriter(buf, maxBatchSizeBytes, false)
		err = jjson.Encode(w, v)
		if err != nil {
			// the user just gets a generic error saying that the json is bad
			c.err = jsonrpc.NewInternalError("server sent bad json")
			// json marshaling errors are reported to the Send call, not the user
			return err
		}
		c.payload = json.RawMessage(bytes.TrimSuffix(buf.Bytes(), []byte{'\n'}))
		return nil
	}
	return nil
}

func (c *batchingRespWriter) Notify(method string, v any) error {
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

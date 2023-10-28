package server

import (
	"bytes"
	"context"
	"net/http"
	"sync"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/util/go/bufpool"
	"github.com/goccy/go-json"
	"golang.org/x/sync/semaphore"
)

// 16mb... should be more than enough for any batch.
// you shouldn't be batching more than this
// TODO: make this configurable
const maxBatchSizeBytes = 1024 * 1024 * 1024 * 16

var _ codec.ResponseWriter = (*callRespWriter)(nil)

// callRespWriter is NOT thread safe
type callRespWriter struct {
	cr  *callResponder
	msg *codec.Message
	ctx context.Context

	noStream bool
	doneMu   *semaphore.Weighted

	payload json.RawMessage
	err     error

	sendCalled bool
	header     http.Header

	mu sync.Mutex
}

func (c *callRespWriter) Send(v any, e error) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.msg.ID == nil {
		return codec.ErrCantSendNotification
	}
	if c.sendCalled {
		return codec.ErrSendAlreadyCalled
	}
	c.sendCalled = true
	// defer the sending of this for later
	if c.doneMu != nil {
		defer c.doneMu.Release(1)
	}
	// batch requests are not individually streamed.
	// the reason is beacuse i couldn't think of a good way to implement it
	// ultimately they need to be buffered. there's some optimistic multiplexing you can
	// do, but that felt really complicated and not worth the time.
	if c.noStream {
		if c.err == nil {
			c.err = e
		}
		if v != nil {
			// json marshaling errors are reported to the handler
			buf := bufpool.GlobalPool.GetStd()
			w := newWriter(buf, maxBatchSizeBytes, false)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				return err
			}
			c.payload = json.RawMessage(bytes.TrimSuffix(buf.Bytes(), []byte{'\n'}))
			return nil
		}
		return nil
	}
	err = c.cr.mu.Acquire(c.ctx, 1)
	if err != nil {
		return err
	}
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}
	defer c.cr.mu.Release(1)
	if c.err != nil {
		e = c.err
	}
	ce := &callEnv{
		err:         e,
		id:          c.msg.ID,
		extrafields: c.msg.ExtraFields,
	}
	if v != nil {
		ce.v = &v
	}
	err = c.cr.send(c.ctx, ce)
	if err != nil {
		return err
	}
	err = c.cr.remote.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (c *callRespWriter) SetExtraField(k string, v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msg.SetExtraField(k, v)
	return nil
}

func (c *callRespWriter) Header() http.Header {
	return c.header
}

func (c *callRespWriter) Notify(method string, v any) error {
	err := c.cr.mu.Acquire(c.ctx, 1)
	if err != nil {
		return err
	}
	defer c.cr.mu.Release(1)
	err = c.cr.notify(c.ctx, &notifyEnv{
		method: method,
		dat:    v,
		extra:  c.msg.ExtraFields,
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

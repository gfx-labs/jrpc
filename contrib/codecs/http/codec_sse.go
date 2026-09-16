package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
	"github.com/gfx-labs/jrpc/pkg/serverutil"
	"github.com/gfx-labs/sse"
)

var _ jsonrpc.ReaderWriter = (*SseCodec)(nil)

// SseCodec is used for subscriptions over http.
// note that every message is buffered multiple times - it's a bit inefficient.
// if you need more efficient streaming of large blobs, consider using a different interface
type SseCodec struct {
	ctx context.Context
	cn  context.CancelFunc

	r    *http.Request
	w    http.ResponseWriter
	i    jsonrpc.PeerInfo
	sink *sse.EventSink
	f    http.Flusher

	msgs jsonrpc.Bundle

	cur bytes.Buffer
}

func NewSseCodec(w http.ResponseWriter, r *http.Request) (*SseCodec, error) {
	// attempt to upgrade
	c := &SseCodec{
		r: r,
		w: w,
		i: jsonrpc.PeerInfo{
			Transport:  "http",
			RemoteAddr: r.RemoteAddr,
			HTTP:       r.Clone(r.Context()),
		},
	}
	c.ctx, c.cn = context.WithCancel(r.Context())
	eventSink, err := sse.DefaultUpgrader.Upgrade(w, r)
	if err != nil {
		return nil, err
	}
	c.sink = eventSink
	method_up := r.URL.Query().Get("method")
	if method_up == "" {
		method_up = strings.TrimPrefix(r.URL.Path, "/")
	}
	params, _ := url.QueryUnescape(r.URL.Query().Get("params"))
	var param []byte
	// try to read params as base64
	if pb, err := base64.URLEncoding.DecodeString(params); err == nil {
		param = pb
	} else {
		// otherwise just take them raw
		param = []byte(params)
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}
	c.msgs = &serverutil.SimpleBundle{
		Msgs: []*jsonrpc.Message{{
			ID:     jsonrpc.NewId(id),
			Method: method_up,
			Params: param,
		}},
		Batch: false,
	}
	return c, nil
}

// gets the peer info
func (c *SseCodec) PeerInfo() jsonrpc.PeerInfo {
	return c.i
}

func (c *SseCodec) ReadBatch(ctx context.Context) (jsonrpc.Bundle, error) {
	if c.msgs == nil {
		return nil, context.Canceled
	}
	defer func() {
		c.msgs = nil
	}()
	return c.msgs, nil
}

// closes the connection
func (c *SseCodec) Write(p []byte) (n int, err error) {
	return c.cur.Write(p)
}

func (c *SseCodec) Flush() error {
	bts := c.cur.Bytes()
	err := c.sink.Encode(&sse.Event{
		Event: []byte("object"),
		Data:  bts,
	})
	c.cur.Reset()
	if err != nil {
		return err
	}
	if c.f != nil {
		c.f.Flush()
	}
	return nil
}

func (c *SseCodec) Close() error {
	if c.f != nil {
		c.f.Flush()
	}
	c.cn()
	return nil
}

// Closed returns a channel which is closed when the connection is closed.
func (c *SseCodec) Closed() <-chan struct{} {
	return c.ctx.Done()
}

// RemoteAddr returns the peer address of the connection.
func (c *SseCodec) RemoteAddr() string {
	return c.r.RemoteAddr
}

package http

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/serverutil"
)

var _ jsonrpc.ReaderWriter = (*HttpCodec)(nil)

type HttpCodec struct {
	ctx context.Context
	cn  context.CancelFunc

	r *http.Request
	w http.ResponseWriter
	i jsonrpc.PeerInfo

	f http.Flusher

	msgs *serverutil.Bundle
}

func NewCodec(w http.ResponseWriter, r *http.Request) (jsonrpc.ReaderWriter, error) {
	switch strings.ToUpper(r.Method) {
	case http.MethodGet:
		if r.Header.Get("Accept") == "text/event-stream" || r.URL.Query().Has("sse") {
			return NewSseCodec(w, r)
		}
		return NewGetCodec(w, r), nil
	case http.MethodPost:
		return NewPostCodec(w, r)
	default:
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return nil, errors.New("method not allowed")
	}
}

func NewGetCodec(w http.ResponseWriter, r *http.Request) *HttpCodec {
	c := &HttpCodec{
		r: r,
		w: w,
		i: jsonrpc.PeerInfo{
			Transport:  "http",
			RemoteAddr: r.RemoteAddr,
			HTTP:       r.Clone(r.Context()),
		},
	}
	c.ctx, c.cn = context.WithCancel(r.Context())
	flusher, ok := w.(http.Flusher)
	if ok {
		c.f = flusher
	}

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
	c.msgs = &serverutil.Bundle{
		Messages: []*jsonrpc.Message{{
			ID:     jsonrpc.NewId(id),
			Method: method_up,
			Params: param,
		}},
		Batch: false,
	}
	return c
}

func NewPostCodec(w http.ResponseWriter, r *http.Request) (*HttpCodec, error) {
	c := &HttpCodec{
		r: r,
		w: w,
		i: jsonrpc.PeerInfo{
			Transport:  "http",
			RemoteAddr: r.RemoteAddr,
			HTTP:       r.Clone(r.Context()),
		},
	}
	c.ctx, c.cn = context.WithCancel(r.Context())
	flusher, ok := w.(http.Flusher)
	if ok {
		c.f = flusher
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	c.msgs = serverutil.ParseBundle(data)

	pathMethod := strings.TrimPrefix(r.URL.Path, "/")
	for _, v := range c.msgs.Messages {
		if v != nil {
			if v.Method == "" {
				v.Method = pathMethod
			}
			if v.ID == nil {
				v.ID = jsonrpc.NewId(1)
			}
		}
	}
	return c, nil
}

// gets the peer info
func (c *HttpCodec) PeerInfo() jsonrpc.PeerInfo {
	return c.i
}

func (c *HttpCodec) ReadBatch(ctx context.Context) ([]*jsonrpc.Message, bool, error) {
	if c.msgs == nil {
		return nil, false, jsonrpc.ErrNoMoreBatches
	}
	defer func() {
		c.msgs = nil
	}()
	return c.msgs.Messages, c.msgs.Batch, nil
}

// closes the connection
func (c *HttpCodec) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *HttpCodec) Flush() error {
	c.w.Write([]byte{'\n'})
	if c.f != nil {
		c.f.Flush()
	}
	return nil
}

func (c *HttpCodec) Close() error {
	if c.f != nil {
		c.f.Flush()
	}
	c.cn()
	return nil
}

// Closed returns a channel which is closed when the connection is closed.
func (c *HttpCodec) Closed() <-chan struct{} {
	return c.ctx.Done()
}

// RemoteAddr returns the peer address of the connection.
func (c *HttpCodec) RemoteAddr() string {
	return c.r.RemoteAddr
}

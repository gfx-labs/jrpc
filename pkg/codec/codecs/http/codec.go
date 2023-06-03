package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/codec"
)

type Codec struct {
	ctx context.Context
	cn  func()

	r    *http.Request
	w    http.ResponseWriter
	msgs chan json.RawMessage
	errs chan error
}

func NewCodec(r *http.Request, w http.ResponseWriter) *Codec {
	ctx, cn := context.WithCancel(r.Context())
	c := &Codec{
		ctx:  ctx,
		cn:   cn,
		r:    r,
		w:    w,
		msgs: make(chan json.RawMessage, 1),
		errs: make(chan error, 1),
	}
	go c.doRead()
	return c
}

// gets the peer info
func (c *Codec) PeerInfo() codec.PeerInfo {
	ci := codec.PeerInfo{
		Transport:  "http",
		RemoteAddr: c.r.RemoteAddr,
		HTTP: codec.HttpInfo{
			Version:   c.r.Proto,
			UserAgent: c.r.UserAgent(),
			Host:      c.r.Host,
			Headers:   c.r.Header.Clone(),
		},
	}
	ci.HTTP.Origin = c.r.Header.Get("X-Real-Ip")
	if ci.HTTP.Origin == "" {
		ci.HTTP.Origin = c.r.Header.Get("X-Forwarded-For")
	}
	if ci.HTTP.Origin == "" {
		ci.HTTP.Origin = c.r.Header.Get("Origin")
	}
	if ci.HTTP.Origin == "" {
		ci.HTTP.Origin = c.r.RemoteAddr
	}
	return ci
}

func (r *Codec) doReadGet() (msgs json.RawMessage, err error) {
	method_up := r.r.URL.Query().Get("method")
	params, _ := url.QueryUnescape(r.r.URL.Query().Get("params"))
	param := []byte(params)
	if pb, err := base64.URLEncoding.DecodeString(params); err == nil {
		param = pb
	}
	id := r.r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}
	req := jrpc.NewRequest(r.ctx, id, method_up, json.RawMessage(param))
	return req.MarshalJSON()
}

var ErrInvalidContentType = errors.New("invalid content type")

func (c *Codec) doRead() {
	contentMatches := true
	types := c.r.Header.Values("content-type")
	for _, v := range types {
		// TODO: check content type
		_ = v
	}
	if !contentMatches {
		c.errs <- ErrInvalidContentType
		return
	}
	var data json.RawMessage
	var err error
	// TODO: implement eventsource
	switch c.r.Method {
	case http.MethodGet:
		data, err = c.doReadGet()
		return
	case http.MethodPost:
		data, err = io.ReadAll(c.r.Body)
	}
	if err != nil {
		c.errs <- err
		return
	}
	c.msgs <- data
}

// json.RawMessage can be an array of requests. if it is, then it is a batch request
func (c *Codec) ReadBatch(ctx context.Context) (msgs json.RawMessage, err error) {
	select {
	case ans := <-c.msgs:
		return ans, nil
	case err := <-c.errs:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

// closes the connection
func (c *Codec) Close() error {
	c.cn()
	return nil
}

func (c *Codec) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

// Closed returns a channel which is closed when the connection is closed.
func (c *Codec) Closed() <-chan struct{} {
	return c.ctx.Done()
}

// RemoteAddr returns the peer address of the connection.
func (c *Codec) RemoteAddr() string {
	return ""
}

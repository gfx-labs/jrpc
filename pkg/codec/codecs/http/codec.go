package http

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/codec"
)

type Codec struct {
	ctx context.Context
	cn  func()

	r     *http.Request
	w     http.ResponseWriter
	wr    *bufio.Writer
	msgs  chan json.RawMessage
	errCh chan httpError

	i codec.PeerInfo
}

type httpError struct {
	code int
	err  error
}

func NewCodec(w http.ResponseWriter, r *http.Request) *Codec {
	c := &Codec{
		r:     r,
		w:     w,
		wr:    bufio.NewWriter(w),
		msgs:  make(chan json.RawMessage, 1),
		errCh: make(chan httpError, 1),
	}
	ctx := r.Context()
	c.ctx, c.cn = context.WithCancel(ctx)
	c.peerInfo()
	c.doRead()
	return c
}
func (c *Codec) peerInfo() {
	c.i.Transport = "http"
	c.i.RemoteAddr = c.r.RemoteAddr
	c.i.HTTP = codec.HttpInfo{
		Version:   c.r.Proto,
		UserAgent: c.r.UserAgent(),
		Host:      c.r.Host,
		Headers:   c.r.Header.Clone(),
	}
	c.i.HTTP.Origin = c.r.Header.Get("X-Real-Ip")
	if c.i.HTTP.Origin == "" {
		c.i.HTTP.Origin = c.r.Header.Get("X-Forwarded-For")
	}
	if c.i.HTTP.Origin == "" {
		c.i.HTTP.Origin = c.r.Header.Get("Origin")
	}
	if c.i.HTTP.Origin == "" {
		c.i.HTTP.Origin = c.r.RemoteAddr
	}
}

// gets the peer info
func (c *Codec) PeerInfo() codec.PeerInfo {
	return c.i
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

// validateRequest returns a non-zero response code and error message if the
// request is invalid.
func ValidateRequest(r *http.Request) (int, error) {
	if r.Method == http.MethodPut || r.Method == http.MethodDelete {
		return http.StatusMethodNotAllowed, errors.New("method not allowed")
	}
	if r.ContentLength > maxRequestContentLength {
		err := fmt.Errorf("content length too large (%d>%d)", r.ContentLength, maxRequestContentLength)
		return http.StatusRequestEntityTooLarge, err
	}
	// Allow OPTIONS (regardless of content-type)
	if r.Method == http.MethodOptions {
		return 0, nil
	}
	// Check content-type
	if mt, _, err := mime.ParseMediaType(r.Header.Get("content-type")); err == nil {
		for _, accepted := range acceptedContentTypes {
			if accepted == mt {
				return 0, nil
			}
		}
	}
	// Invalid content-type ignored for now
	return 0, nil
	//err := fmt.Errorf("invalid content type, only %s is supported", contentType)
	//return http.StatusUnsupportedMediaType, err
}

func (c *Codec) doRead() {
	code, err := ValidateRequest(c.r)
	if err != nil {
		c.errCh <- httpError{
			code: code,
			err:  err,
		}
		return
	}
	go func() {
		var data json.RawMessage
		// TODO: implement eventsource
		switch c.r.Method {
		case http.MethodGet:
			data, err = c.doReadGet()
		case http.MethodPost:
			data, err = io.ReadAll(c.r.Body)
		}
		if err != nil {
			c.errCh <- httpError{
				code: http.StatusInternalServerError,
				err:  err,
			}
			return
		}
		c.msgs <- data
	}()
}

// json.RawMessage can be an array of requests. if it is, then it is a batch request
func (c *Codec) ReadBatch(ctx context.Context) (msgs json.RawMessage, err error) {
	select {
	case ans := <-c.msgs:
		return ans, nil
	case err := <-c.errCh:
		http.Error(c.w, err.err.Error(), err.code)
		return nil, err.err
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
	return c.wr.Write(p)
}

func (c *Codec) Flush() (err error) {
	err = c.wr.Flush()
	if err != nil {
		return err
	}
	c.cn()
	return
}

// Closed returns a channel which is closed when the connection is closed.
func (c *Codec) Closed() <-chan struct{} {
	return c.ctx.Done()
}

// RemoteAddr returns the peer address of the connection.
func (c *Codec) RemoteAddr() string {
	return c.r.RemoteAddr
}

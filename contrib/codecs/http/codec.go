package http

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/serverutil"
)

var _ jsonrpc.ReaderWriter = (*Codec)(nil)

// Reusable codec. use Reset()
type Codec struct {
	ctx context.Context
	cn  func()

	r     *http.Request
	w     http.ResponseWriter
	wr    *bufio.Writer
	msgs  chan *serverutil.Bundle
	errCh chan httpError

	mu sync.Mutex

	i jsonrpc.PeerInfo
}

type httpError struct {
	code int
	err  error
}

func NewCodec(w http.ResponseWriter, r *http.Request) *Codec {
	c := &Codec{}
	c.Reset(w, r)
	return c
}

func (c *Codec) Reset(w http.ResponseWriter, r *http.Request) {
	c.wr = bufio.NewWriter(w)
	if w == nil {
		c.wr = bufio.NewWriter(io.Discard)
	}
	c.r = r
	c.w = w
	c.msgs = make(chan *serverutil.Bundle, 1)
	c.errCh = make(chan httpError, 1)

	ctx := c.r.Context()
	c.ctx, c.cn = context.WithCancel(ctx)
	c.peerInfo()
	c.doRead()
}

func (c *Codec) peerInfo() {
	c.i.Transport = "http"
	c.i.RemoteAddr = c.r.RemoteAddr
	c.i.HTTP = jsonrpc.HttpInfo{
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
func (c *Codec) PeerInfo() jsonrpc.PeerInfo {
	return c.i
}

func (r *Codec) doReadGet() (msg *serverutil.Bundle, err error) {
	method_up := r.r.URL.Query().Get("method")
	if method_up == "" {
		method_up = r.r.URL.Path
	}
	params, _ := url.QueryUnescape(r.r.URL.Query().Get("params"))
	param := []byte(params)
	if pb, err := base64.URLEncoding.DecodeString(params); err == nil {
		param = pb
	}
	id := r.r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}
	return &serverutil.Bundle{
		Messages: []*jsonrpc.Message{{
			ID:     jsonrpc.NewId(id),
			Method: method_up,
			Params: param,
		}},
		Batch: false,
	}, nil
}

func (r *Codec) doReadRPC() (msg *serverutil.Bundle, err error) {
	method_up := r.r.URL.Query().Get("method")
	if method_up == "" {
		method_up = r.r.URL.Path
	}
	id := r.r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}
	data, err := io.ReadAll(r.r.Body)
	if err != nil {
		return nil, err
	}
	return &serverutil.Bundle{
		Messages: []*jsonrpc.Message{{
			ID:     jsonrpc.NewId(id),
			Method: method_up,
			Params: data,
		}},
		Batch: false,
	}, nil
}

func (r *Codec) doReadPost() (msg *serverutil.Bundle, err error) {
	data, err := io.ReadAll(r.r.Body)
	if err != nil {
		return nil, err
	}
	return serverutil.ParseBundle(data), nil
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
		var data *serverutil.Bundle
		// TODO: implement eventsource
		switch strings.ToUpper(c.r.Method) {
		case http.MethodGet:
			data, err = c.doReadGet()
		case "RPC":
			data, err = c.doReadRPC()
		case http.MethodPost:
			data, err = c.doReadPost()
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

func (c *Codec) ReadBatch(ctx context.Context) ([]*jsonrpc.Message, bool, error) {
	select {
	case ans := <-c.msgs:
		return ans.Messages, ans.Batch, nil
	case err := <-c.errCh:
		http.Error(c.w, err.err.Error(), err.code)
		return nil, false, err.err
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case <-c.ctx.Done():
		return nil, false, c.ctx.Err()
	}
}

// closes the connection
func (c *Codec) Write(p []byte) (n int, err error) {
	return c.wr.Write(p)
}

func (c *Codec) Flush() error {
	defer c.cn()
	err := c.wr.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (c *Codec) Close() error {
	c.cn()
	return nil
}

// Closed returns a channel which is closed when the connection is closed.
func (c *Codec) Closed() <-chan struct{} {
	return c.ctx.Done()
}

// RemoteAddr returns the peer address of the connection.
func (c *Codec) RemoteAddr() string {
	return c.r.RemoteAddr
}

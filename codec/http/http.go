// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package jrpc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"

	"gfx.cafe/open/jrpc/codec"
	"gfx.cafe/util/go/bufpool"

	json "github.com/goccy/go-json"
)

const (
	maxRequestContentLength = 1024 * 1024 * 5
	contentType             = "application/json"
)

// https://www.jsonrpc.org/historical/json-rpc-over-http.html#id13
var acceptedContentTypes = []string{
	// https://www.jsonrpc.org/historical/json-rpc-over-http.html#id13
	contentType, "application/json-rpc", "application/jsonrequest",
	// these are added because they make sense, fight me!
	"application/jsonrpc2", "application/json-rpc2", "application/jrpc",
}

// HTTPTimeouts represents the configuration params for the HTTP RPC server.
type HTTPTimeouts struct {
	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration
}

// DefaultHTTPTimeouts represents the default timeout values used if further
// configuration is not provided.
var DefaultHTTPTimeouts = HTTPTimeouts{
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	IdleTimeout:  120 * time.Second,
}

// httpServerConn turns a HTTP connection into a Conn.
type httpServerConn struct {
	io.Reader
	io.Writer

	jc codec.ReaderWriter

	r *http.Request
	w http.ResponseWriter

	pi codec.PeerInfo
}

func newHTTPServerConn(r *http.Request, w http.ResponseWriter, pi codec.PeerInfo) codec.ReaderWriter {
	c := &httpServerConn{Writer: w, r: r, pi: pi}
	// if the request is a GET request, and the body is empty, we turn the request into fake json rpc request, see below
	// https://www.jsonrpc.org/historical/json-rpc-over-http.html#encoded-parameters
	// we however allow for non base64 encoded parameters to be passed
	if r.Method == http.MethodGet {
		// default id 1
		id := `1`
		id_up := r.URL.Query().Get("id")
		if id_up != "" {
			id = id_up
		}
		method_up := r.URL.Query().Get("method")
		params, _ := url.QueryUnescape(r.URL.Query().Get("params"))
		param := []byte(params)
		if pb, err := base64.URLEncoding.DecodeString(params); err == nil {
			param = pb
		}
		buf := bufpool.GetStd()
		json.NewEncoder(buf).Encode(jsonrpcMessage{
			ID:     NewStringIDPtr(id),
			Method: method_up,
			Params: param,
		})
		c.Reader = buf
	} else {
		// it's a post request or whatever, so just process it like normal
		c.Reader = io.LimitReader(r.Body, maxRequestContentLength)
	}
	c.jc = NewCodec(c)
	return c
}

func (c *httpServerConn) PeerInfo() PeerInfo {
	return c.pi
}

func (c *httpServerConn) ReadBatch() (messages []*jsonrpcMessage, batch bool, err error) {
	return c.jc.ReadBatch()
}

func (c *httpServerConn) WriteJSON(ctx context.Context, v any) error {
	return c.jc.WriteJSON(ctx, v)
}

func (c *httpServerConn) Close() error {
	return nil
}

// Closed returns a channel which will be closed when Close is called
func (c *httpServerConn) Closed() <-chan any {
	return c.jc.Closed()
}

// RemoteAddr returns the peer address of the underlying connection.
func (t *httpServerConn) RemoteAddr() string {
	return t.PeerInfo().RemoteAddr
}

// SetWriteDeadline does nothing and always returns nil.
func (t *httpServerConn) SetWriteDeadline(time.Time) error { return nil }

// ServeHTTP serves JSON-RPC requests over HTTP.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Permit dumb empty requests for remote health-checks (AWS)
	if r.Method == http.MethodGet && r.ContentLength == 0 && r.URL.RawQuery == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if code, err := validateRequest(r); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	// Create request-scoped context.
	connInfo := PeerInfo{
		Transport:  "http",
		RemoteAddr: r.RemoteAddr,
		HTTP: HttpInfo{
			Version:      r.Proto,
			UserAgent:    r.UserAgent(),
			Host:         r.Host,
			Headers:      r.Header.Clone(),
			WriteHeaders: w.Header(),
		},
	}
	connInfo.HTTP.Version = r.Proto
	connInfo.HTTP.Host = r.Host
	connInfo.HTTP.Origin = r.Header.Get("X-Real-Ip")
	if connInfo.HTTP.Origin == "" {
		connInfo.HTTP.Origin = r.Header.Get("X-Forwarded-For")
	}
	if connInfo.HTTP.Origin == "" {
		connInfo.HTTP.Origin = r.Header.Get("Origin")
	}
	if connInfo.HTTP.Origin == "" {
		connInfo.HTTP.Origin = r.RemoteAddr
	}
	// the headers used
	connInfo.HTTP.Headers = r.Header

	ctx := r.Context()
	ctx = context.WithValue(ctx, peerInfoContextKey{}, connInfo)

	// All checks passed, create a codec that reads directly from the request body
	// until EOF, writes the response to w, and orders the server to process a
	// single request.
	w.Header().Set("content-type", contentType)

	codec := newHTTPServerConn(r, w, connInfo)
	defer codec.Close()
	s.serveSingleRequest(ctx, codec)
}

// validateRequest returns a non-zero response code and error message if the
// request is invalid.
func validateRequest(r *http.Request) (int, error) {
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

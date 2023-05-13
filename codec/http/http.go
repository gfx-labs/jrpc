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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/codec"
	"gfx.cafe/util/go/bufpool"
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
type requestCodec struct {
	r *http.Request
	w http.ResponseWriter

	ctx context.Context
	cn  func()

	requestBuffer *bytes.Buffer
	pi            codec.PeerInfo
}

func NewRequestCodec(r *http.Request, w http.ResponseWriter) *requestCodec {
	// Create request-scoped context.
	connInfo := codec.PeerInfo{
		Transport:  "http",
		RemoteAddr: r.RemoteAddr,
		HTTP: codec.HttpInfo{
			Version:   r.Proto,
			UserAgent: r.UserAgent(),
			Host:      r.Host,
			Headers:   r.Header.Clone(),
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
	buf := bufpool.GetStd()

	ctx, cn := context.WithCancel(r.Context())

	return &requestCodec{
		ctx:           ctx,
		cn:            cn,
		r:             r,
		w:             w,
		pi:            connInfo,
		requestBuffer: buf,
	}

}

// gets the peer info
func (r *requestCodec) PeerInfo() codec.PeerInfo {
	return r.pi
}

// json.RawMessage can be an array of requests. if it is, then it is a batch request
func (r *requestCodec) ReadBatch(ctx context.Context) (msgs json.RawMessage, err error) {
	if r.r.Method == http.MethodGet {
		return r.readBatchGet(ctx)
	}
	if r.r.Method == http.MethodPost {
		return r.readBatch(ctx)
	}
	return nil, fmt.Errorf("invalid request")
}

func (r *requestCodec) readBatchGet(ctx context.Context) (msgs json.RawMessage, err error) {
	method_up := r.r.URL.Query().Get("method")
	params, _ := url.QueryUnescape(r.r.URL.Query().Get("params"))
	param := []byte(params)
	if pb, err := base64.URLEncoding.DecodeString(params); err == nil {
		param = pb
	}
	req := jrpc.NewRequestInt(ctx, 1, method_up, json.RawMessage(param))
	return req.MarshalJSON()
}

func (r *requestCodec) readBatch(ctx context.Context) (msgs json.RawMessage, err error) {
	rd := io.LimitReader(r.r.Body, maxRequestContentLength)
	_, err = io.Copy(r.requestBuffer, rd)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.ctx.Done():
		return nil, r.ctx.Err()
	}
	return json.RawMessage(r.requestBuffer.Bytes()), nil
}

// closes the connection
func (r *requestCodec) Close() error {
	r.cn()
	bufpool.PutStd(r.requestBuffer)
	return nil
}

func (r *requestCodec) Write(p []byte) (n int, err error) {
	return r.w.Write(p)
}

// Closed returns a channel which is closed when the connection is closed.
func (r *requestCodec) Closed() <-chan struct{} {
	return r.r.Context().Done()
}

// RemoteAddr returns the peer address of the connection.
func (r *requestCodec) RemoteAddr() string {
	return r.pi.RemoteAddr
}

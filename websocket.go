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
	"net/http"
	"net/url"
	"sync"
	"time"

	"gfx.cafe/open/jrpc/wsjson"
	"git.tuxpa.in/a/zlog/log"
	"nhooyr.io/websocket"
)

const (
	wsReadBuffer       = 1024
	wsWriteBuffer      = 1024
	wsPingInterval     = 60 * time.Second
	wsPingWriteTimeout = 5 * time.Second
	wsPongTimeout      = 30 * time.Second
	wsMessageSizeLimit = 32 * 1024 * 1024
)

var wsBufferPool = new(sync.Pool)

// WebsocketHandler returns a handler that serves JSON-RPC to WebSocket connections.
//
// allowedOrigins should be a comma-separated list of allowed origin URLs.
// To allow connections with any origin, pass "*".
func (s *Server) WebsocketHandler(allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns:       allowedOrigins,
			CompressionMode:      websocket.CompressionContextTakeover,
			CompressionThreshold: 512,
		})
		if err != nil {
			log.Debug().Err(err).Msg("WebSocket upgrade failed")
			return
		}
		codec := newWebsocketCodec(r.Context(), conn, r.Host, r.Header)
		s.ServeCodec(codec)
	})
}

type wsHandshakeError struct {
	err    error
	status string
}

func (e wsHandshakeError) Error() string {
	s := e.err.Error()
	if e.status != "" {
		s += " (HTTP status " + e.status + ")"
	}
	return s
}

// that is listening on the given endpoint using the provided dialer.
func DialWebsocketWithDialer(ctx context.Context, endpoint, origin string, opts *websocket.DialOptions) (*Client, error) {
	endpoint, header, err := wsClientHeaders(endpoint, origin)
	if err != nil {
		return nil, err
	}
	opts.HTTPHeader = header
	return newClient(ctx, func(ctx context.Context) (ServerCodec, error) {
		conn, resp, err := websocket.Dial(ctx, endpoint, opts)
		if err != nil {
			hErr := wsHandshakeError{err: err}
			if resp != nil {
				hErr.status = resp.Status
			}
			return nil, hErr
		}
		out := newWebsocketCodec(resp.Request.Context(), conn, endpoint, header)
		return out, err
	})
}

// DialWebsocket creates a new RPC client that communicates with a JSON-RPC server
// that is listening on the given endpoint.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
func DialWebsocket(ctx context.Context, endpoint, origin string) (*Client, error) {
	endpoint, header, err := wsClientHeaders(endpoint, origin)
	if err != nil {
		return nil, err
	}
	dialer := &websocket.DialOptions{
		CompressionMode:      websocket.CompressionContextTakeover,
		CompressionThreshold: 512,
		HTTPHeader:           header,
	}
	return DialWebsocketWithDialer(ctx, endpoint, origin, dialer)
}

func wsClientHeaders(endpoint, origin string) (string, http.Header, error) {
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return endpoint, nil, err
	}
	header := make(http.Header)
	if origin != "" {
		header.Add("origin", origin)
	}
	if endpointURL.User != nil {
		b64auth := base64.StdEncoding.EncodeToString([]byte(endpointURL.User.String()))
		header.Add("authorization", "Basic "+b64auth)
		endpointURL.User = nil
	}
	return endpointURL.String(), header, nil
}

type websocketCodec struct {
	*jsonCodec
	conn *websocket.Conn
	info PeerInfo

	wg        sync.WaitGroup
	pingReset chan struct{}
}

func heartbeat(ctx context.Context, c *websocket.Conn, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		err := c.Ping(ctx)
		if err != nil {
			return
		}

		t.Reset(time.Minute)
	}
}

func newWebsocketCodec(ctx context.Context, c *websocket.Conn, host string, req http.Header) ServerCodec {
	c.SetReadLimit(wsMessageSizeLimit)
	jsonWriter := func(v any) error {
		return wsjson.Write(context.Background(), c, v)
	}
	jsonReader := func(v any) error {
		return wsjson.Read(context.Background(), c, v)
	}
	conn := websocket.NetConn(ctx, c, websocket.MessageText)
	wc := &websocketCodec{
		jsonCodec: NewFuncCodec(conn, jsonWriter, jsonReader).(*jsonCodec),
		conn:      c,
		pingReset: make(chan struct{}, 1),
		info: PeerInfo{
			Transport: "ws",
		},
	}
	// Fill in connection details.
	wc.info.HTTP.Host = host
	// traefik proxy protocol headers
	wc.info.HTTP.Origin = req.Get("X-Real-Ip")
	if wc.info.HTTP.Origin == "" {
		wc.info.HTTP.Origin = req.Get("X-Forwarded-For")
	}
	// origin header fallback
	if wc.info.HTTP.Origin == "" {
		wc.info.HTTP.Origin = req.Get("origin")
	}
	wc.info.RemoteAddr = wc.info.HTTP.Origin
	wc.info.HTTP.UserAgent = req.Get("User-Agent")
	wc.info.HTTP.Headers = req
	// Start pinger.
	go heartbeat(ctx, c, wsPingInterval)
	return wc
}

func (wc *websocketCodec) close() {
	wc.jsonCodec.close()
	wc.conn.CloseRead(context.Background())
}

func (wc *websocketCodec) peerInfo() PeerInfo {
	return wc.info
}

func (wc *websocketCodec) writeJSON(ctx context.Context, v any) error {
	err := wc.jsonCodec.writeJSON(ctx, v)
	if err == nil {
		// Notify pingLoop to delay the next idle ping.
		select {
		case wc.pingReset <- struct{}{}:
		default:
		}
	}
	return err
}

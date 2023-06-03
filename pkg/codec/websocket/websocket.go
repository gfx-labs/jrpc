package websocket

import (
	"context"
	"encoding/base64"
	"encoding/json"
	codec2 "gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/codec/websocket/wsjson"
	"net/http"
	"net/url"
	"time"

	"gfx.cafe/open/jrpc"
	"nhooyr.io/websocket"
)

// WebsocketHandler returns a handler that serves JSON-RPC to WebSocket connections.
//
// allowedOrigins should be a comma-separated list of allowed origin URLs.
// To allow connections with any origin, pass "*".
func WebsocketHandler(s *jrpc.Server, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns:       allowedOrigins,
			CompressionMode:      websocket.CompressionContextTakeover,
			CompressionThreshold: 4096,
		})
		if err != nil {
			return
		}
		codec := newWebsocketCodec(r.Context(), conn, r.Host, r.Header)
		s.ServeCodec(codec)
	})
}

func NewHandshakeError(err error, status string) error {
	return &wsHandshakeError{err, status}
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

func WsClientHeaders(endpoint, origin string) (string, http.Header, error) {
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
	conn *websocket.Conn
	info codec2.PeerInfo

	pingReset chan struct{}

	closed chan any
}

// if there is more than one message, it is a batch request
func (w *websocketCodec) ReadBatch(ctx context.Context) (msgs json.RawMessage, err error) {
	w.conn.SetReadLimit(WsMessageSizeLimit)
	err = wsjson.Read(ctx, w.conn, &msgs)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// Closed returns a channel which is closed when the connection is closed.
func (w *websocketCodec) Closed() <-chan any {
	return w.closed
}

// RemoteAddr returns the peer address of the connection.
func (w *websocketCodec) RemoteAddr() string {
	return w.info.RemoteAddr
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

func newWebsocketCodec(ctx context.Context, c *websocket.Conn, host string, req http.Header) codec2.ReaderWriter {
	wc := &websocketCodec{
		conn:      c,
		pingReset: make(chan struct{}, 1),
		info: codec2.PeerInfo{
			Transport: "ws",
		},
		closed: make(chan any),
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
	go heartbeat(ctx, c, WsPingInterval)
	return wc
}

func (wc *websocketCodec) Close() error {
	wc.conn.CloseRead(context.Background())
	close(wc.closed)
	return nil
}

func (wc *websocketCodec) PeerInfo() codec2.PeerInfo {
	return wc.info
}

func (wc *websocketCodec) WriteJSON(ctx context.Context, v any) error {
	err := wsjson.Write(ctx, wc.conn, v)
	if err == nil {
		// Notify pingLoop to delay the next idle ping.
		select {
		case wc.pingReset <- struct{}{}:
		default:
		}
	}
	return err
}

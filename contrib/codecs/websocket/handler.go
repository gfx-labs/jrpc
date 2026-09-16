package websocket

import (
	"net/http"

	"github.com/coder/websocket"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
	"github.com/gfx-labs/jrpc/pkg/server"
)

type Server struct {
	Handler jsonrpc.Handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Handler == nil {
		http.Error(w, "no server set", http.StatusInternalServerError)
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := newWebsocketCodec(r.Context(), conn, "", r)
	err = server.ServeCodec(r.Context(), c, s.Handler)
	if err != nil {
		//slog.Error("codec err", "error", err)
	}
}

// WebsocketHandler returns a handler that serves JSON-RPC to WebSocket connections.
//
// allowedOrigins should be a comma-separated list of allowed origin URLs.
// To allow connections with any origin, pass "*".
func WebsocketHandler(s jsonrpc.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns:       allowedOrigins,
			CompressionMode:      websocket.CompressionContextTakeover,
			CompressionThreshold: 4096,
		})
		if err != nil {
			return
		}
		codec := newWebsocketCodec(r.Context(), conn, r.Host, r)
		err = server.ServeCodec(r.Context(), codec, s)
		if err != nil {
			// slog.Error("codec err", "error", err)
		}
	})
}

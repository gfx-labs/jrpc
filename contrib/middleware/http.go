package middleware

import (
	"path"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

func HttpPathMethodPrefix(h jsonrpc.Handler) jsonrpc.Handler {
	return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		if r.Peer.HTTP != nil {
			r.Method = path.Join(r.Peer.HTTP.URL.Path, r.Method)
		}
		h.ServeRPC(w, r)
	})
}

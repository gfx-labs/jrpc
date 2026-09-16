package middleware

import (
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

func EmptyMethodInvalid(next jsonrpc.Handler) jsonrpc.Handler {
	return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		if len(r.Method) == 0 {
			w.Send(nil, jsonrpc.NewInvalidRequestError("invalid request"))
			return
		}
		next.ServeRPC(w, r)
	})
}

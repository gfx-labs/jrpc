package middleware

import (
	"strings"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

var LegacyUnderscoreReplacer = MethodReplacer(strings.NewReplacer("_", "/"))

// MethodReplacer will use the replacer on every method before handling
func MethodReplacer(replacer *strings.Replacer) jsonrpc.Middleware {
	return func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			r.Method = replacer.Replace(r.Method)
			next.ServeRPC(w, r)
		})
	}
}

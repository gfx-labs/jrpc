package middleware

import (
	"strings"

	"gfx.cafe/open/jrpc/pkg/codec"
)

var LegacyUnderscoreReplacer = MethodReplacer(strings.NewReplacer("_", "/"))

// MethodReplacer will use the replacer on every method before handling
func MethodReplacer(replacer *strings.Replacer) codec.Middleware {
	return func(next codec.Handler) codec.Handler {
		return codec.HandlerFunc(func(w codec.ResponseWriter, r *codec.Request) {
			r.Method = replacer.Replace(r.Method)
			next.ServeRPC(w, r)
		})
	}
}

package middleware

// Ported from Goji's middleware, source:
// jrpcs://github.com/zenazn/goji/tree/master/web/middleware

import (
	"time"

	"gfx.cafe/open/jrpc"
	"git.tuxpa.in/a/zlog/log"
)

func Logger(next jrpc.Handler) jrpc.Handler {
	fn := func(w jrpc.ResponseWriter, r *jrpc.Request) {
		start := time.Now()
		next.ServeRPC(w, r)
		log.Trace().
			Stringer("time", time.Since(start)).
			Str("remote", r.Remote()).
			Str("method", r.Method()).
			Str("params", string(r.Msg().Params)).Msg("RPC Request")
	}
	return jrpc.HandlerFunc(fn)
}

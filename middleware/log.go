package middleware

import (
	"context"
	"time"

	"gfx.cafe/open/jrpc"
	"tuxpa.in/a/zlog"
	"tuxpa.in/a/zlog/log"
)

// Key to use when setting the request ID.
type ctxKeyLogger int

// RequestIDKey is the key that holds the unique request ID in a request context.
const LoggerKey ctxKeyLogger = 76

func Logger(next jrpc.Handler) jrpc.Handler {
	fn := func(w jrpc.ResponseWriter, r *jrpc.Request) {
		start := time.Now()
		l := log.Trace().
			Str("id", GetReqID(r.Context())).
			Str("remote", r.Remote()).
			Str("method", r.Method).
			Str("params", string(r.Msg().Params))
		next.ServeRPC(w, r.WithContext(context.WithValue(r.Context(), LoggerKey, l)))
		l = l.Stringer("dur", time.Since(start))
		l.Msg("RPC Request")
	}
	return jrpc.HandlerFunc(fn)
}

func GetLogger(ctx context.Context) *zlog.Event {
	if ctx == nil {
		return log.Debug()
	}
	if lgr, ok := ctx.Value(LoggerKey).(*zlog.Event); ok {
		return lgr
	}
	return log.Debug()
}

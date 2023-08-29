package middleware

import (
	"context"
	"time"

	"gfx.cafe/open/jrpc/pkg/codec"

	"log/slog"
)

// Key to use when setting the request ID.
type ctxKeyLogger int

// RequestIDKey is the key that holds the unique request ID in a request context.
const LoggerKey ctxKeyLogger = 76

func NewLogger(logger *slog.Logger) func(next codec.Handler) codec.Handler {
	return func(next codec.Handler) codec.Handler {
		fn := func(w codec.ResponseWriter, r *codec.Request) {
			start := time.Now()
			lg := logger.With(
				"remote", r.Remote(),
				"method", r.Method,
				"params", string(r.Msg().Params),
			)
			if id := GetReqID(r.Context()); id != "" {
				lg = logger.With(
					"req_id", id,
				)
			}
			next.ServeRPC(w, r.WithContext(context.WithValue(r.Context(), LoggerKey, lg)))
			lg = logger.With(
				"params", time.Since(start),
			)
			logger.LogAttrs(r.Context(), slog.LevelDebug, "RPC Request")
		}
		return codec.HandlerFunc(fn)
	}
}

func Logger(next codec.Handler) codec.Handler {
	lh := slog.Default()
	return NewLogger(lh)(next)
}

func GetLogger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if lgr, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return lgr
	}
	return slog.Default()
}

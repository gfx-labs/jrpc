package middleware

import (
	"context"
	"time"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

func Timeout(dur time.Duration) func(h jsonrpc.Handler) jsonrpc.Handler {
	return func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			done := make(chan struct{})
			ctx, cn := context.WithTimeout(r.Context(), dur)
			r = r.WithContext(ctx)
			go func() {
				next.ServeRPC(w, r)
				close(done)
				cn()
			}()
			<-ctx.Done()
			cn()
			<-done
		})
	}
}

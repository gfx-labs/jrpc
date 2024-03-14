package http

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
)

func JrpcToHttp(h jsonrpc.Handler) http.Handler {
	return HttpHandler(h)
}

func HttpHandler(h jsonrpc.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h == nil {
			http.Error(w, "no server set", http.StatusInternalServerError)
			return
		}
		c, err := NewCodec(w, r)
		if err != nil {
			return
		}
		w.Header().Set("content-type", contentType)
		err = server.ServeCodec(r.Context(), c, h)
		if err != nil && !errors.Is(err, context.Canceled) {
			//  slog.Error("codec err", "err", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		<-c.Closed()
	}), &http2.Server{})
}

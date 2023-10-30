package http

import (
	"net/http"
	"sync"

	"gfx.cafe/open/jrpc/pkg/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func HttpHandler(s *server.Server) http.Handler {
	return h2c.NewHandler(&Server{Server: s}, &http2.Server{})
}

type Server struct {
	Server *server.Server
}

var codecPool = sync.Pool{
	New: func() any {
		return &Codec{}
	},
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Server == nil {
		http.Error(w, "no server set", http.StatusInternalServerError)
		return
	}
	c := NewCodec(w, r)
	w.Header().Set("content-type", contentType)
	err := s.Server.ServeCodec(r.Context(), c)
	if err != nil {
		//	slog.Error("codec err", "err", err)
	}
	http.Error(w, "Internal Error", http.StatusInternalServerError)
	<-c.Closed()
}

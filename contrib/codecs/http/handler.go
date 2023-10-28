package http

import (
	"net/http"
	"sync"

	"gfx.cafe/open/jrpc/pkg/server"
)

func HttpHandler(s *server.Server) http.Handler {
	return &Server{Server: s}
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
	<-c.Closed()
}

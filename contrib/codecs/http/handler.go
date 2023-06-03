package http

import (
	"gfx.cafe/open/jrpc/pkg/server"
	"net/http"
)

type Server struct {
	Server *server.Server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Server == nil {
		http.Error(w, "no server set", http.StatusInternalServerError)
		return
	}
	c := NewCodec(w, r)
	s.Server.ServeCodec(r.Context(), c)
}

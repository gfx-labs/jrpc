package http

import (
	"net/http"

	"gfx.cafe/open/jrpc"
)

type Server struct {
	Server *jrpc.Server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Server == nil {
		http.Error(w, "no server set", http.StatusInternalServerError)
		return
	}
	c := NewCodec(w, r)
	s.Server.ServeCodec(r.Context(), c)
}

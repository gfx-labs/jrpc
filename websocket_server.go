package jrpc

import (
	"net/http"
	"strings"
)

type WebsocketServer struct {
	s *Server
}

func (s *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if IsWebsocket(r) {
		s.s.WebsocketHandler([]string{"*"}).ServeHTTP(w, r)
		return
	}
	s.s.ServeHTTP(w, r)
}

func IsWebsocket(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func (s *Server) ServeHTTPWithWss(cb func(w http.ResponseWriter, r *http.Request) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cb != nil {
			if cb(w, r) {
				return
			}
		}
		if IsWebsocket(r) {
			s.WebsocketHandler([]string{"*"}).ServeHTTP(w, r)
			return
		}
		s.ServeHTTP(w, r)
	})
}

package http

import (
	"net/http/httptest"

	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
)

func ServerMaker() (*server.Server, jrpctest.ClientMaker, func()) {
	s := jrpctest.NewServer()
	hsrv := httptest.NewServer(HttpHandler(s))
	return s, func() jsonrpc.Conn {
		conn, err := DialHTTP(hsrv.URL)
		if err != nil {
			panic(err)
		}
		return conn
	}, hsrv.Close
}

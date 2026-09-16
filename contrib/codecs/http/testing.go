package http

import (
	"net/http/httptest"

	"github.com/gfx-labs/jrpc/internal/jrpctest"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

func ServerMaker() (jsonrpc.Handler, jrpctest.ClientMaker, func()) {
	s := jrpctest.NewRouter()
	hsrv := httptest.NewServer(HttpHandler(s))
	return s, func() jsonrpc.Conn {
		conn, err := DialHTTP(hsrv.URL)
		if err != nil {
			panic(err)
		}
		return conn
	}, hsrv.Close
}

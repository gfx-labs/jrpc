package websocket

import (
	"context"
	"net/http/httptest"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/server"
)

func ServerMaker() (*server.Server, jrpctest.ClientMaker, func()) {
	s := jrpctest.NewServer()
	hsrv := httptest.NewServer(&Server{Server: s})
	return s, func() codec.Conn {
		conn, err := DialWebsocket(context.Background(), hsrv.URL, "")
		if err != nil {
			panic(err)
		}
		return conn
	}, hsrv.Close
}

package websocket

import (
	"context"
	"net/http/httptest"

	"gfx.cafe/open/jrpc/internal/jrpctest"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

func ServerMaker() (jsonrpc.Handler, jrpctest.ClientMaker, func()) {
	s := jrpctest.NewRouter()
	hsrv := httptest.NewServer(&Server{Handler: s})
	return s, func() jsonrpc.Conn {
		conn, err := DialWebsocket(context.Background(), hsrv.URL, "")
		if err != nil {
			panic(err)
		}
		return conn
	}, hsrv.Close
}

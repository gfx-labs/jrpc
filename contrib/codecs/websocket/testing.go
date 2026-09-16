package websocket

import (
	"context"
	"net/http/httptest"

	"github.com/gfx-labs/jrpc/internal/jrpctest"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
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

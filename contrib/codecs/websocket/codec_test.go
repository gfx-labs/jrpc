package websocket

import (
	"context"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/server"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: func() (*server.Server, jrpctest.ClientMaker, func()) {
			s := jrpctest.NewServer()
			hsrv := httptest.NewServer(&Server{Server: s})
			return s, func() codec.Conn {
				conn, err := DialWebsocket(context.Background(), hsrv.URL, "")
				require.NoError(t, err)
				return conn
			}, hsrv.Close
		},
	})
}

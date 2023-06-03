package http

import (
	"net/http/httptest"
	"testing"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"github.com/stretchr/testify/require"
)

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: func() (*jrpc.Server, jrpctest.ClientMaker, func()) {
			s := jrpctest.NewServer()
			hsrv := httptest.NewServer(&Server{Server: s})
			return s, func() jrpc.Conn {
				conn, err := DialHTTP(hsrv.URL)
				require.NoError(t, err)
				return conn
			}, hsrv.Close
		},
	})
}

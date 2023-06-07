package rdwr_test

import (
	"context"
	"io"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"

	"gfx.cafe/open/jrpc/pkg/jrpctest"
)

func TestBasicSuite(t *testing.T) {

	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: func() (*server.Server, jrpctest.ClientMaker, func()) {
			rd_s, wr_s := io.Pipe()
			rd_c, wr_c := io.Pipe()
			s := jrpctest.NewServer()
			clientCodec := rdwr.NewCodec(rd_c, wr_s)
			go func() {
				s.ServeCodec(context.Background(), clientCodec)
			}()
			return s, func() codec.Conn {
				return rdwr.NewClient(rd_s, wr_c, nil)
			}, func() {}
		},
	})
}

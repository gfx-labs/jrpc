package inproc_test

import (
	"context"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/inproc"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"

	"gfx.cafe/open/jrpc/pkg/jrpctest"
)

func mockServerMaker() (*server.Server, jrpctest.ClientMaker, func()) {
	s := jrpctest.NewServer()
	clientCodec := inproc.NewCodec()
	go func() {
		s.ServeCodec(context.Background(), clientCodec)
	}()
	return s, func() codec.Conn {
		return inproc.NewClient(clientCodec, nil)
	}, func() {}
}

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: mockServerMaker,
	})
}

package inproc

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/server"
)

func ServerMaker() (*server.Server, jrpctest.ClientMaker, func()) {
	s := jrpctest.NewServer()
	clientCodec := NewCodec()
	go func() {
		s.ServeCodec(context.Background(), clientCodec)
	}()
	return s, func() codec.Conn {
		return NewClient(clientCodec, nil)
	}, func() {}
}

package rdwr

import (
	"context"
	"io"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/server"
)

func ServerMaker() (*server.Server, jrpctest.ClientMaker, func()) {
	rd_s, wr_s := io.Pipe()
	rd_c, wr_c := io.Pipe()
	s := jrpctest.NewServer()
	clientCodec := NewCodec(rd_c, wr_s)
	go func() {
		s.ServeCodec(context.Background(), clientCodec)
	}()
	return s, func() codec.Conn {
		return NewClient(rd_s, wr_c)
	}, func() {}
}

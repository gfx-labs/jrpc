package rdwr

import (
	"context"
	"io"

	"gfx.cafe/open/jrpc/internal/jrpctest"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
)

func ServerMaker() (jsonrpc.Handler, jrpctest.ClientMaker, func()) {
	rd_s, wr_s := io.Pipe()
	rd_c, wr_c := io.Pipe()
	s := jrpctest.NewRouter()
	clientCodec := NewCodec(rd_c, wr_s)
	ctx, cn := context.WithCancel(context.Background())
	go func() {
		server.ServeCodec(ctx, clientCodec, s)
	}()
	return s, func() jsonrpc.Conn {
			return NewClient(rd_s, wr_c)
		}, func() {
			cn()
		}
}

package broker

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/server"
)

type Server struct {
	Server *server.Server
}

func (s *Server) ServeSpoke(ctx context.Context, stream ServerSpoke) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		req, fn, err := stream.ReadRequest(ctx)
		if err != nil {
			continue
		}
		if req == nil {
			continue
		}
		cd := NewCodec(req, fn)
		s.Server.ServeCodec(ctx, cd)
	}
}

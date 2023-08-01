package redis

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/server"
	"tuxpa.in/a/zlog/log"
)

type Server struct {
	Server *server.Server
}

func (s *Server) ServeRedis(ctx context.Context, stream *ServerStream) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		req, fn, err := stream.ReadRequest(ctx)
		if err != nil {
			log.Err(err).Msg("while reading bpop")
			continue
		}
		if req == nil {
			continue
		}
		cd := NewCodec(req, fn)
		s.Server.ServeCodec(ctx, cd)
	}
}

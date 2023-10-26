package broker

import (
	"context"
	"log/slog"

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
		go func() {
			err := s.Server.ServeCodec(ctx, cd)
			if err != nil {
				slog.Error("codec err", "err", err)
			}
			cd.Close()
		}()
	}
}

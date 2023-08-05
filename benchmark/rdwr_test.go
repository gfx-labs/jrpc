package benchmark

import (
	"context"
	"io"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
)

func BenchmarkIoPipeClientServer(b *testing.B) {
	ctx := context.Background()
	s := jrpctest.NewServer()

	b.Run("SingleClient", func(b *testing.B) {
		rd_s, wr_s := io.Pipe()
		rd_c, wr_c := io.Pipe()
		clientCodec := rdwr.NewCodec(rd_c, wr_s, nil)
		go s.ServeCodec(ctx, clientCodec)
		conn := rdwr.NewClient(rd_s, wr_c)
		for i := 0; i < b.N; i++ {
			conn.Do(ctx, nil, "ping", nil)
		}
	})
}

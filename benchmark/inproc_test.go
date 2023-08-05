package benchmark

import (
	"context"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/inproc"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
)

func BenchmarkInprocClientServer(b *testing.B) {
	ctx := context.Background()
	s := jrpctest.NewServer()

	b.Run("SingleClient", func(b *testing.B) {
		clientCodec := inproc.NewCodec()
		go s.ServeCodec(ctx, clientCodec)
		conn := inproc.NewClient(clientCodec, nil)
		for i := 0; i < b.N; i++ {
			conn.Do(ctx, nil, "ping", nil)
		}
	})
}

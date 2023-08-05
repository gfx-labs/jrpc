package benchmark

import (
	"context"
	"net/http/httptest"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"github.com/stretchr/testify/require"
)

func BenchmarkHttpClientServer(b *testing.B) {
	ctx := context.Background()
	s := jrpctest.NewServer()

	b.Run("SingleClient", func(b *testing.B) {
		hsrv := httptest.NewServer(&http.Server{Server: s})
		conn, err := http.DialHTTP(hsrv.URL)
		require.NoError(b, err)
		for i := 0; i < b.N; i++ {
			conn.Do(ctx, nil, "ping", nil)
		}
	})
}

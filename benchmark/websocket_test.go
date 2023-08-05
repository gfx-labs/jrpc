package benchmark

import (
	"context"
	"net/http/httptest"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"github.com/stretchr/testify/require"
)

func BenchmarkWebsocketClientServer(b *testing.B) {
	ctx := context.Background()
	s := jrpctest.NewServer()

	b.Run("SingleClient", func(b *testing.B) {
		hsrv := httptest.NewServer(&websocket.Server{Server: s})
		conn, err := websocket.DialWebsocket(ctx, hsrv.URL, "")
		require.NoError(b, err)
		for i := 0; i < b.N; i++ {
			conn.Do(ctx, nil, "ping", nil)
		}
	})
}

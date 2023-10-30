package benchmark

import (
	"context"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/server"
)

func TestBenchmarkSuite(t *testing.T) {
	var executeTest = jrpctest.TestExecutor(rdwr.ServerMaker)
	var makeTest = func(name string, fm jrpctest.TestContext) {
		t.Run(name, func(t *testing.T) {
			executeTest(t, fm)
		})
	}

	ctx := context.Background()
	makeTest("SingleClient", func(t *testing.T, server *server.Server, client codec.Conn) {
		err := client.Do(ctx, nil, "test_ping", nil)
		if err != nil {
			t.Error(err)
		}
	})
}

func runBenchmarkSuite(b *testing.B, sm jrpctest.ServerMaker) {
	ctx := context.Background()
	executeBench := jrpctest.BenchExecutor(sm)
	var makeBench = func(name string, fm jrpctest.BenchContext) {
		b.Run(name, func(b *testing.B) {
			executeBench(b, fm)
		})
	}
	makeBench("SingleClient", func(b *testing.B, server *server.Server, client codec.Conn) {
		for i := 0; i < b.N; i++ {
			err := client.Do(ctx, nil, "test_ping", nil)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkSimpleSuite(b *testing.B) {

	makers := map[string]jrpctest.ServerMaker{
		"Http":      http.ServerMaker,
		"WebSocket": websocket.ServerMaker,
		"IoPipe":    rdwr.ServerMaker,
	}

	for k, v := range makers {
		b.Run(k, func(b *testing.B) {
			runBenchmarkSuite(b, v)
		})
	}

}

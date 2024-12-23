package benchmark

import (
	"context"
	"sync"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/http"
	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/contrib/codecs/websocket"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type testCase struct {
	name     string
	method   string
	parallel bool
}

var testCases = []testCase{
	{"SingleClient", "/small/largeResp", false},
	{"SingleClientMedium", "/medium/largeResp", false},
	{"SingleClientLarge", "/large/largeResp", false},
	{"ParallelClient", "/small/largeResp", true},
	{"ParallelClientMedium", "/medium/largeResp", true},
	{"ParallelClientLarge", "/large/largeResp", true},
}

func runTestCase(ctx context.Context, client jsonrpc.Conn, method string, parallel bool) error {
	if !parallel {
		return client.Do(ctx, nil, method, nil)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 16)

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := client.Do(ctx, nil, method, nil); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		return err // Return first error encountered
	}
	return nil
}

func TestBenchmarkSuite(t *testing.T) {
	executeTest := jrpctest.TestExecutor(rdwr.ServerMaker)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executeTest(t, func(t *testing.T, h jsonrpc.Handler, client jsonrpc.Conn) {
				if err := runTestCase(ctx, client, tc.method, tc.parallel); err != nil {
					t.Error(err)
				}
			})
		})
	}
}

func runBenchmarkSuite(b *testing.B, sm jrpctest.ServerMaker) {
	ctx := context.Background()
	executeBench := jrpctest.BenchExecutor(sm)

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			executeBench(b, func(b *testing.B, h jsonrpc.Handler, client jsonrpc.Conn) {
				for i := 0; i < b.N; i++ {
					if err := runTestCase(ctx, client, tc.method, tc.parallel); err != nil {
						panic(err)
					}
				}
			})
		})
	}
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

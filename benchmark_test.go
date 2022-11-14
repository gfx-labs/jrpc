package jrpc

import (
	"testing"

	"golang.org/x/sync/errgroup"
)

func BenchmarkClientHTTPEcho(b *testing.B) {
	server := newTestServer()
	defer server.Stop()
	client, hs := httpTestClient(server, "http", nil)
	defer hs.Close()
	defer client.Close()

	// Launch concurrent requests.
	wantBack := map[string]any{
		"one": map[string]any{"two": "three"},
		"e":   map[string]any{"two": "three"},
		"oe":  map[string]any{"two": "three"},
		"on":  map[string]any{"two": "three"},
	}

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		eg := &errgroup.Group{}
		eg.SetLimit(4)
		for i := 0; i < 1000; i++ {
			eg.Go(func() error {
				return client.Call(nil, "test_echoAny", []any{1, 2, 3, 4, 56, 6, wantBack, wantBack, wantBack})
			})
		}
		eg.Wait()
	}
}
func BenchmarkClientHTTPEchoEmpty(b *testing.B) {
	server := newTestServer()
	defer server.Stop()
	client, hs := httpTestClient(server, "http", nil)
	defer hs.Close()
	defer client.Close()

	// Launch concurrent requests.

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		eg := &errgroup.Group{}
		eg.SetLimit(4)
		for i := 0; i < 1000; i++ {
			eg.Go(func() error {
				return client.Call(nil, "test_echoAny", 0)
			})
		}
		eg.Wait()
	}
}
func BenchmarkClientWebsocketEcho(b *testing.B) {
	server := newTestServer()
	defer server.Stop()
	client, hs := httpTestClient(server, "ws", nil)
	defer hs.Close()
	defer client.Close()

	// Launch concurrent requests.
	wantBack := map[string]any{
		"one": map[string]any{"two": "three"},
		"e":   map[string]any{"two": "three"},
		"oe":  map[string]any{"two": "three"},
		"on":  map[string]any{"two": "three"},
	}

	payload := []any{1, 2, 3, 4, 56, 6, wantBack, wantBack, wantBack}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		eg := &errgroup.Group{}
		eg.SetLimit(4)
		for i := 0; i < 1000; i++ {
			eg.Go(func() error {
				return client.Call(nil, "test_echoAny", payload)
			})
		}
		eg.Wait()
	}

}
func BenchmarkClientWebsocketEchoEmpty(b *testing.B) {
	server := newTestServer()
	defer server.Stop()
	client, hs := httpTestClient(server, "ws", nil)
	defer hs.Close()
	defer client.Close()

	// Launch concurrent requests.
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		eg := &errgroup.Group{}
		eg.SetLimit(4)
		for i := 0; i < 1000; i++ {
			eg.Go(func() error {
				return client.Call(nil, "test_echoAny", 0)
			})
		}
		eg.Wait()
	}
}

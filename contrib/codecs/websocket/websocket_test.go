package websocket_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gfx-labs/jrpc/contrib/codecs/websocket"
	"github.com/gfx-labs/jrpc/contrib/jmux"
	"github.com/gfx-labs/jrpc/internal/jrpctest"
)

func TestWebsocketClientHeaders(t *testing.T) {
	t.Parallel()

	endpoint, header, err := websocket.WsClientHeaders("wss://testuser:test-PASS_01@example.com:1234", "https://example.com")
	if err != nil {
		t.Fatalf("wsGetConfig failed: %s", err)
	}
	if endpoint != "wss://example.com:1234" {
		t.Fatal("User should have been stripped from the URL")
	}
	if header.Get("authorization") != "Basic dGVzdHVzZXI6dGVzdC1QQVNTXzAx" {
		t.Fatal("Basic auth header is incorrect")
	}
	if header.Get("origin") != "https://example.com" {
		t.Fatal("Origin not set")
	}
}

// This test checks that the server rejects connections from disallowed origins.
func TestWebsocketOriginCheck(t *testing.T) {
	t.Parallel()

	var (
		srv     = jrpctest.NewRouter()
		httpsrv = httptest.NewServer(websocket.WebsocketHandler(srv, []string{"http://example.com"}))
		wsURL   = "ws:" + strings.TrimPrefix(httpsrv.URL, "http:")
	)
	defer httpsrv.Close()

	client, err := websocket.DialWebsocket(context.Background(), wsURL, "http://ekzample.com")
	if err == nil {
		client.Close()
		t.Fatal("no error for wrong origin")
	}
	wantErr := "expected handshake response status code 101 but got 403"
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("wrong error for wrong origin: got: '%q', want: '%s'", err, wantErr)
	}

	// Connections without origin header should work.
	client, err = websocket.DialWebsocket(context.Background(), wsURL, "")
	if err != nil {
		t.Fatalf("error for empty origin: %v", err)
	}
	client.Close()
}

// This test checks whether calls exceeding the request size limit are rejected.
func TestWebsocketLargeCall(t *testing.T) {
	t.Parallel()

	var (
		srv     = jrpctest.NewRouter()
		httpsrv = httptest.NewServer(websocket.WebsocketHandler(srv, []string{"*"}))
		wsURL   = "ws:" + strings.TrimPrefix(httpsrv.URL, "http:")
	)
	defer httpsrv.Close()

	client, err := websocket.DialWebsocket(context.Background(), wsURL, "")
	if err != nil {
		t.Fatalf("can't dial: %v", err)
	}
	defer client.Close()

	// This call sends slightly less than the limit and should work.
	var result jrpctest.EchoResult
	arg := strings.Repeat("x", websocket.MaxRequestContentLength-200)
	if err := client.Do(nil, &result, "test_echo", []any{arg, 1}); err != nil {
		t.Fatalf("valid call didn't work: %v", err)
	}
	if result.String != arg {
		t.Fatal("wrong string echoed")
	}

	// This call sends twice the allowed size and shouldn't work.
	arg = strings.Repeat("x", websocket.MaxRequestContentLength*2)
	err = client.Do(nil, &result, "test_echo", []any{arg})
	if err == nil {
		t.Fatal("no error for too large call")
	}
}

// This checks that the websocket transport can deal with large messages.
func TestClientWebsocketLargeMessage(t *testing.T) {
	mux := jmux.NewMux()
	var (
		srv     = mux
		httpsrv = httptest.NewServer(websocket.WebsocketHandler(srv, nil))
		wsURL   = "ws:" + strings.TrimPrefix(httpsrv.URL, "http:")
	)
	defer httpsrv.Close()

	respLength := websocket.WsMessageSizeLimit - 50
	if err := mux.RegisterStruct("test", jrpctest.LargeRespService{Length: respLength}); err != nil {
		t.Fatal(err)
	}

	c, err := websocket.DialWebsocket(context.Background(), wsURL, "")
	if err != nil {
		t.Fatal(err)
	}

	var r string
	if err := c.Do(nil, &r, "test/largeResp", nil); err != nil {
		t.Fatal("call failed:", err)
	}
	if len(r) != respLength {
		t.Fatalf("response has wrong length %d, want %d", len(r), respLength)
	}
}

// wsPingTestServer runs a WebSocket server which accepts a single subscription request.
// When a value arrives on sendPing, the server sends a ping frame, waits for a matching
// pong and finally delivers a single subscription result.
//func wsPingTestServer(t *testing.T, sendPing <-chan struct{}) *http.Server {
//	var srv http.Server
//	shutdown := make(chan struct{})
//	srv.RegisterOnShutdown(func() {
//		close(shutdown)
//	})
//	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
//			OriginPatterns: []string{"*"},
//		})
//		// Upgrade to WebSocket.
//		if err != nil {
//			t.Errorf("server WS upgrade error: %v", err)
//			return
//		}
//		defer conn.Close(websocket.StatusAbnormalClosure, "closed")
//
//		// Handle the connection.
//		wsPingTestHandler(t, conn, shutdown, sendPing)
//	})
//
//	// Start the server.
//	listener, err := net.Listen("tcp", "127.0.0.1:0")
//	if err != nil {
//		t.Fatal("can't listen:", err)
//	}
//	srv.Addr = listener.Addr().String()
//	go srv.Serve(listener)
//	return &srv
//}

///func wsPingTestHandler(t *testing.T, conn *websocket.Conn, shutdown, sendPing <-chan struct{}) {
///	// Canned responses for the eth_subscribe call in TestClientWebsocketPing.
///	const (
///		subResp   = `{"jsonrpc":"2.0","id":1,"result":"0x00"}`
///		subNotify = `{"jsonrpc":"2.0","method":"eth_subscription","params":{"subscription":"0x00","result":1}}`
///	)
///
///	// Handle subscribe request.
///	if _, _, err := conn.ReadMessage(); err != nil {
///		t.Errorf("server read error: %v", err)
///		return
///	}
///
///	if err := conn.WriteMessage(websocket.TextMessage, []byte(subResp)); err != nil {
///		t.Errorf("server write error: %v", err)
///		return
///	}
///
///	// Read from the connection to process control messages.
///	var pongCh = make(chan string)
///	conn.SetPongHandler(func(d string) error {
///		t.Logf("server got pong: %q", d)
///		pongCh <- d
///		return nil
///	})
///	go func() {
///		for {
///			typ, msg, err := conn.ReadMessage()
///			if err != nil {
///				return
///			}
///			t.Logf("server got message (%d): %q", typ, msg)
///		}
///	}()
///
///	// Write messages.
///	var (
///		wantPong string
///		timer    = time.NewTimer(0)
///	)
///	defer timer.Stop()
///	<-timer.C
///	for {
///		select {
///		case _, open := <-sendPing:
///			if !open {
///				sendPing = nil
///			}
///			t.Logf("server sending ping")
///			conn.WriteMessage(websocket.PingMessage, []byte("ping"))
///			wantPong = "ping"
///		case data := <-pongCh:
///			if wantPong == "" {
///				t.Errorf("unexpected pong")
///			} else if data != wantPong {
///				t.Errorf("got pong with wrong data %q", data)
///			}
///			wantPong = ""
///			timer.Reset(200 * time.Millisecond)
///		case <-timer.C:
///			t.Logf("server sending response")
///			conn.WriteMessage(websocket.TextMessage, []byte(subNotify))
///		case <-shutdown:
///			conn.Close()
///			return
///		}
///	}
///}

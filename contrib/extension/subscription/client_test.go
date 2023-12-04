package subscription

import (
	"context"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
)

func TestWrapClient(t *testing.T) {
	engine := NewEngine()
	r := jmux.NewRouter()
	r.Use(engine.Middleware())
	r.HandleFunc("echo", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		_ = w.Send(r.Params, nil)
	})
	// extremely fast subscription to fill buffers to get a higher chance that we receive another message while trying
	// to unsubscribe
	r.HandleFunc("test/subscribe", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		notifier, ok := NotifierFromContext(r.Context())
		if !ok {
			_ = w.Send(nil, ErrNotificationsUnsupported)
			return
		}
		go func() {
			idx := 0
			for {
				select {
				case <-r.Context().Done():
					return
				case <-notifier.Err():
					return
				default:
				}
				_ = notifier.Notify(idx)
				idx += 1
			}
		}()
	})
	srv := server.NewServer(r)
	handler := codecs.WebsocketHandler(srv, []string{"*"})
	httpSrv := http.Server{
		Addr:    ":8855",
		Handler: handler,
	}
	listener, err := net.Listen("tcp", ":8855")
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		if err := httpSrv.Serve(listener); err != nil {
			t.Error(err)
			return
		}
	}()

	cl, err := UpgradeConn(jrpc.Dial("ws://localhost:8855"))
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		var res string
		if err = cl.Do(context.Background(), &res, "echo", "test"); err != nil {
			t.Error(err)
			return
		}
		if res != "test" {
			t.Errorf(`expected "test" but got %#v`, res)
			return
		}

		ch := make(chan int, 1)
		sub, err := cl.Subscribe(context.Background(), "test", ch, nil)
		if err != nil {
			t.Error(err)
			return
		}

		go func() {
			time.Sleep(2 * time.Second)
			_ = sub.Unsubscribe()
		}()

		func() {
			for {
				select {
				case err, ok := <-sub.Err():
					if ok {
						t.Errorf("sub errored: %v", err)
					}
					return
				case v := <-ch:
					log.Printf("%v", v)
				}
			}
		}()
	}
}

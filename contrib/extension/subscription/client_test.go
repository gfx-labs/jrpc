package subscription

import (
	"context"
	"net/http/httptest"
	_ "net/http/pprof"
	"strings"
	"testing"
	"time"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
)

func TestSubscription(t *testing.T) {
	const count = 100

	engine := NewEngine()
	r := jmux.NewRouter()
	r.Use(engine.Middleware())
	r.HandleFunc("test/subscribe", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		notifier, ok := NotifierFromContext(r.Context())
		if !ok {
			_ = w.Send(nil, ErrNotificationsUnsupported)
			return
		}

		go func() {
			time.Sleep(10 * time.Millisecond)
			for i := 0; i < count; i++ {
				if err := notifier.Notify(i); err != nil {
					panic(err)
				}
			}
		}()
	})

	srv := server.NewServer(r)
	defer srv.Shutdown(context.Background())
	handler := codecs.WebsocketHandler(srv, []string{"*"})
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	wsURL := "ws:" + strings.TrimPrefix(httpSrv.URL, "http:")
	cl, err := UpgradeConn(jrpc.Dial(wsURL))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err = cl.Close(); err != nil {
			t.Error(err)
		}
	}()

	ch := make(chan int, count)
	sub, err := cl.Subscribe(context.Background(), "test", ch, nil)
	defer func() {
		if err = sub.Unsubscribe(); err != nil {
			t.Error(err)
		}
	}()

	for i := 0; i < count; i++ {
		v := <-ch
		if v != i {
			t.Errorf("expected %d but got %d", i, v)
		}
	}
}

func TestUnsubscribeNoRead(t *testing.T) {
	engine := NewEngine()
	r := jmux.NewRouter()
	r.Use(engine.Middleware())
	r.HandleFunc("test/subscribe", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		notifier, ok := NotifierFromContext(r.Context())
		if !ok {
			_ = w.Send(nil, ErrNotificationsUnsupported)
			return
		}

		go func() {
			time.Sleep(10 * time.Millisecond)
			for i := 0; i < 10; i++ {
				if err := notifier.Notify(i); err != nil {
					panic(err)
				}
			}
		}()
	})

	srv := server.NewServer(r)
	defer srv.Shutdown(context.Background())
	handler := codecs.WebsocketHandler(srv, []string{"*"})
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	wsURL := "ws:" + strings.TrimPrefix(httpSrv.URL, "http:")
	cl, err := UpgradeConn(jrpc.Dial(wsURL))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err = cl.Close(); err != nil {
			t.Error(err)
		}
	}()

	ch := make(chan int)
	sub, err := cl.Subscribe(context.Background(), "test", ch, nil)
	time.Sleep(time.Second)
	if err = sub.Unsubscribe(); err != nil {
		t.Error(err)
		return
	}
}

func TestWrapClient(t *testing.T) {
	engine := NewEngine()
	r := jmux.NewRouter()
	r.Use(engine.Middleware())
	r.HandleFunc("echo", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		err := w.Send(r.Params, nil)
		if err != nil {
			t.Error(err)
		}
	})
	// extremely fast subscription to fill buffers to get a higher chance that we receive another message while trying
	// to unsubscribe
	r.HandleFunc("test/subscribe", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		notifier, ok := NotifierFromContext(r.Context())
		if !ok {
			err := w.Send(nil, ErrNotificationsUnsupported)
			if err != nil {
				t.Error(err)
			}
			return
		}
		go func() {
			time.Sleep(10 * time.Millisecond)
			idx := 0
			for {
				select {
				case <-r.Context().Done():
					return
				case <-notifier.Err():
					return
				default:
				}
				err := notifier.Notify(idx)
				if err != nil {
					t.Error(err)
				}
				idx += 1
			}
		}()
	})
	srv := server.NewServer(r)
	defer srv.Shutdown(context.Background())
	handler := codecs.WebsocketHandler(srv, []string{"*"})
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	wsURL := "ws:" + strings.TrimPrefix(httpSrv.URL, "http:")
	cl, err := UpgradeConn(jrpc.Dial(wsURL))
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err = cl.Close(); err != nil {
			t.Error(err)
		}
	}()

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

		ch := make(chan int, 101)
		var sub ClientSubscription
		sub, err = cl.Subscribe(context.Background(), "test", ch, nil)
		if err != nil {
			t.Error(err)
			return
		}

		func() {
			for {
				select {
				case err, ok := <-sub.Err():
					if ok {
						t.Errorf("sub errored: %v", err)
					}
					return
				case n, ok := <-ch:
					if !ok {
						return
					}
					if n == 100 {
						if err = sub.Unsubscribe(); err != nil {
							t.Error(err)
							return
						}
					}
				}
			}
		}()
	}
}

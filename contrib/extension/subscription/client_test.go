package subscription

import (
	"context"
	"encoding/json"
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

func newRouter(t *testing.T) jmux.Router {
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
		var count int
		_ = json.Unmarshal(r.Params, &count)
		go func() {
			time.Sleep(10 * time.Millisecond)
			for idx := 0; count == 0 || idx < count; idx++ {
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
			}
		}()
	})

	return r
}

func newServer(t *testing.T) (Conn, func()) {
	r := newRouter(t)
	srv := server.NewServer(r)
	handler := codecs.WebsocketHandler(srv, []string{"*"})
	httpSrv := httptest.NewServer(handler)

	wsURL := "ws:" + strings.TrimPrefix(httpSrv.URL, "http:")
	cl, err := UpgradeConn(jrpc.Dial(wsURL))
	if err != nil {
		t.Error(err)
		return nil, nil
	}

	return cl, func() {
		_ = cl.Close()
		httpSrv.Close()
		srv.Shutdown(context.Background())
	}
}

func TestSubscription(t *testing.T) {
	const count = 100

	cl, done := newServer(t)
	defer done()

	ch := make(chan int, count)
	sub, err := cl.Subscribe(context.Background(), "test", ch, count)
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
	cl, done := newServer(t)
	defer done()

	ch := make(chan int)
	sub, err := cl.Subscribe(context.Background(), "test", ch, 10)
	time.Sleep(time.Second)
	if err = sub.Unsubscribe(); err != nil {
		t.Error(err)
		return
	}
}

func TestWrapClient(t *testing.T) {
	cl, done := newServer(t)
	defer done()

	for i := 0; i < 10; i++ {
		var res string
		if err := cl.Do(context.Background(), &res, "echo", "test"); err != nil {
			t.Error(err)
			return
		}
		if res != "test" {
			t.Errorf(`expected "test" but got %#v`, res)
			return
		}

		ch := make(chan int, 101)
		sub, err := cl.Subscribe(context.Background(), "test", ch, nil)
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

func TestCloseClient(t *testing.T) {
	cl, done := newServer(t)
	defer done()

	ch := make(chan int)
	sub, err := cl.Subscribe(context.Background(), "test", ch, nil)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		if err := cl.Close(); err != nil {
			t.Error(err)
		}
	}()

	for {
		select {
		case err, ok := <-sub.Err():
			if ok {
				t.Errorf("sub errored: %v", err)
			}
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
		}
	}
}

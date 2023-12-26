package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/contrib/extension/subscription"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/contrib/middleware"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"

	"gfx.cafe/open/jrpc"
)

func main() {
	engine := subscription.NewEngine()
	r := jmux.NewRouter()
	r.Use(middleware.Logger)
	srv := server.NewServer(r)
	r.HandleFunc("echo", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(r.Params, nil)
	})
	r.Group(func(r jmux.Router) {
		r.Use(engine.Middleware())
		r.HandleFunc("testservice/subscribe", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			notifier, ok := subscription.NotifierFromContext(r.Context())
			if !ok {
				w.Send(nil, subscription.ErrNotificationsUnsupported)
				return
			}
			idx := 0
			for {
				select {
				case <-r.Context().Done():
					return
				case <-notifier.Err():
					return
				default:
				}
				notifier.Notify(idx)
				time.Sleep(1 * time.Second)
				idx = idx + 1
			}
		})
	})
	go func() {
		err := client()
		if err != nil {
			panic(err)
		}
	}()
	log.Println("running on 8855")

	handler := codecs.HttpWebsocketHandler(srv, []string{"*"})
	err := http.ListenAndServe(":8855", handler)
	if err != nil {
		log.Println(err)
	}
}

func client() error {
	cl, err := subscription.UpgradeConn(jrpc.Dial("ws://localhost:8855"))
	if err != nil {
		return err
	}
	out := make(chan int, 1)
	jcs, err := cl.Subscribe(context.TODO(), "testservice", out, []any{"swag"})
	if err != nil {
		return err
	}
	go func() {
		log.Println(<-jcs.Err())
	}()
	defer jcs.Unsubscribe()
	for {
		log.Println("receiving", <-out)
	}
}

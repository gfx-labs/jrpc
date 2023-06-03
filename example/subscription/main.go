package main

import (
	"context"
	"gfx.cafe/open/jrpc/contrib/middleware"
	"log"
	"net/http"
	"time"

	"gfx.cafe/open/jrpc"
)

func main() {

	r := jrpc.NewRouter()
	r.Use(middleware.Logger)
	srv := jrpc.NewServer(r)

	r.HandleFunc("echo", func(w jrpc.ResponseWriter, r *jrpc.Request) {
		w.Send(r.Params(), nil)
	})

	r.HandleFunc("testservice_subscribe", func(w jrpc.ResponseWriter, r *jrpc.Request) {
		sub, err := jrpc.UpgradeToSubscription(w, r)
		w.Send(sub, err)
		if err != nil {
			return
		}
		go func() {
			idx := 0
			for {
				log.Println("sending:", idx)
				err := w.Notify(idx)
				if err != nil {
					return
				}
				time.Sleep(1 * time.Second)
				idx = idx + 1
			}
		}()
	})

	go func() {
		err := client()
		if err != nil {
			panic(err)
		}
	}()
	log.Println("running on 8855")
	log.Println(http.ListenAndServe(":8855", srv.ServeHTTPWithWss(nil)))
}

func client() error {
	cl, err := jrpc.Dial("ws://localhost:8855")
	if err != nil {
		return err
	}
	out := make(chan int, 1)
	jcs, err := cl.Subscribe(context.TODO(), "testservice", out, "swag")
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

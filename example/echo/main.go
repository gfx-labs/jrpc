package main

import (
	"log"
	"net/http"

	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
)

func main() {

	r := jmux.NewRouter()
	srv := server.NewServer(r)

	r.HandleFunc("echo", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(r.Params, nil)
	})

	log.Println("running on 8855")
	log.Println(http.ListenAndServe(":8855", codecs.HttpHandler(srv)))
}

// http://localhost:8855/?method=echo&params=[1,2,3]

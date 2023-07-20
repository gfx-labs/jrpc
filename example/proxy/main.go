package main

import (
	"encoding/json"
	"log"
	"net/http"

	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/contrib/middleware"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"

	"gfx.cafe/open/jrpc"
)

func main() {

	r := jmux.NewRouter()
	r.Use(middleware.Logger)
	c, err := jrpc.Dial("wss://mainnet.rpc.gfx.xyz")
	if err != nil {
		panic(err)
	}

	r.HandleFunc("eth_*", func(w codec.ResponseWriter, r *codec.Request) {
		var res json.RawMessage
		err = c.Do(r.Context(), &res, r.Method, json.RawMessage(r.Params))
		w.Send(res, err)
	})

	log.Println("running on 8855")

	srv := server.NewServer(r)
	log.Println(http.ListenAndServe(":8855", codecs.HttpHandler(srv)))
}

// http://localhost:8855/?method=eth_blockNumber

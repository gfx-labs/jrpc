package main

import (
	"encoding/json"
	"gfx.cafe/open/jrpc/contrib/middleware"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"
	"log"
	"net/http"

	"gfx.cafe/open/jrpc"
)

func main() {

	r := jrpc.NewRouter()
	r.Use(middleware.Logger)
	c, err := jrpc.Dial("wss://mainnet.rpc.gfx.xyz")
	if err != nil {
		panic(err)
	}

	r.HandleFunc("eth_*", func(w codec.ResponseWriter, r *codec.Request) {
		var res json.RawMessage
		err := c.Call(&res, r.Method(), r.ParamSlice()...)
		w.Send(res, err)
	})

	log.Println("running on 8855")

	srv := server.NewServer(r)
	log.Println(http.ListenAndServe(":8855", srv))
}

// http://localhost:8855/?method=eth_blockNumber

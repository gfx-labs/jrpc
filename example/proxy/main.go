package main

import (
	"encoding/json"
	"log"
	"net/http"

	"gfx.cafe/open/jrpc"
)

func main() {

	r := jrpc.NewRouter()
	srv := jrpc.NewServer(r)

	c, err := jrpc.Dial("wss://mainnet.rpc.gfx.xyz")
	if err != nil {
		panic(err)
	}

	r.HandleFunc("eth_*", func(w jrpc.ResponseWriter, r *jrpc.Request) {
		var res json.RawMessage
		err := c.Call(&res, r.Method(), r.ParamSlice()...)
		w.Send(res, err)
	})

	log.Println("running on 8855")
	log.Println(http.ListenAndServe(":8855", srv))
}

// http://localhost:8855/?method=echo&params=[1,2,3]

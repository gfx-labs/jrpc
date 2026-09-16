package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gfx-labs/jrpc/contrib/codecs"
	"github.com/gfx-labs/jrpc/contrib/jmux"
	"github.com/gfx-labs/jrpc/contrib/middleware"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"

	"github.com/gfx-labs/jrpc"
)

func main() {

	r := jmux.NewRouter()
	r.Use(middleware.Logger)
	c, err := jrpc.Dial("wss://mainnet.rpc.gfx.xyz")
	if err != nil {
		panic(err)
	}

	r.HandleFunc("eth_*", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		var res json.RawMessage
		err = c.Do(r.Context(), &res, r.Method, json.RawMessage(r.Params))
		w.Send(res, err)
	})

	log.Println("running on 8855")

	log.Println(http.ListenAndServe(":8855", codecs.HttpHandler(r)))
}

// http://localhost:8855/?method=eth_blockNumber

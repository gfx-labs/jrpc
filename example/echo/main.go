package main

import (
	"log"
	"net/http"

	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

func main() {

	r := jmux.NewRouter()

	r.HandleFunc("echo", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(r.Params, nil)
	})

	r.RegisterFunc("echo/register", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(r.Params, nil)
	})

	server := &EchoServer{}
	r.RegisterStruct("server", server)

	log.Println("running on 8855")
	log.Println(http.ListenAndServe(":8855", codecs.HttpHandler(r)))
}

type EchoServer struct {
}

func (e *EchoServer) Echo(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
	w.Send(r.Params, nil)
}

// http://localhost:8855/?method=echo&params=[1,2,3]
// http://localhost:8855/?method=server/echo&params=[1,2,3]

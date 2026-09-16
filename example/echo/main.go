package main

import (
	"log"
	"net/http"

	"github.com/gfx-labs/jrpc/contrib/codecs"
	"github.com/gfx-labs/jrpc/contrib/jmux"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
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
	log.Println(http.ListenAndServe(":8855", codecs.HttpWebsocketHandler(r, []string{"*"})))
}

type EchoServer struct {
}

func (e *EchoServer) Echo(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
	w.Send(r.Params, nil)
}

// http://localhost:8855/?method=echo&params=[1,2,3]
// http://localhost:8855/?method=server/echo&params=[1,2,3]

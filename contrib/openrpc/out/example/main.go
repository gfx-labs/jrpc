package main

import (
	"context"
	"log"
	"net/http"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/contrib/openrpc/out"
)

func main() {
	handler := &out.RpcHandler{
		FnEthChainId: func(context.Context) (out.Uint, error) { return "0", nil },
	}
	rpc_handler := &out.GoOpenRPCHandler{Srv: handler}
	rmux := jmux.NewMux()
	rpc_handler.RouteRPC(rmux)
	srv := jrpc.NewServer(rmux)
	log.Println("listening on port 9545")
	log.Println(http.ListenAndServe(":9545", codecs.HttpHandler(srv)))
}

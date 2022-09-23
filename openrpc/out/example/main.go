package main

import (
	"context"
	"log"
	"net/http"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/openrpc/out"
)

func main() {
	handler := &out.RpcHandler{
		FnEthChainId: func(context.Context) (out.Uint, error) { return "0", nil },
	}
	rpc_handler := &out.GoOpenRPCHandler{Srv: handler}
	rmux := jrpc.NewMux()
	rpc_handler.RouteRPC(rmux)
	srv := jrpc.NewServer(rmux)
	log.Println(srv.Router().Routes())
	log.Println("listening on port 9545")
	log.Println(http.ListenAndServe(":9545", srv))
}

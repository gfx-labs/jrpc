package main

import (
	"context"
	"gfx.cafe/open/jrpc/contrib/openrpc/out"
	jrpc2 "gfx.cafe/open/jrpc/pkg/server"
	"log"
	"net/http"

	"gfx.cafe/open/jrpc"
)

func main() {
	handler := &out.RpcHandler{
		FnEthChainId: func(context.Context) (out.Uint, error) { return "0", nil },
	}
	rpc_handler := &out.GoOpenRPCHandler{Srv: handler}
	rmux := jrpc.NewMux()
	rpc_handler.RouteRPC(rmux)
	srv := jrpc2.NewServer(rmux)
	log.Println(srv.Router().Routes())
	log.Println("listening on port 9545")
	log.Println(http.ListenAndServe(":9545", srv))
}

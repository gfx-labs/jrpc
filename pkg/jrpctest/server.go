package jrpctest

import (
	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/jmux"
)

func NewTestServer() *jrpc.Server {
	mux := jmux.NewRouter()
	server := jrpc.NewServer(mux)
	mux.HandleFunc("testservice_subscribe", func(w jrpc.ResponseWriter, r *jrpc.Request) {
		sub, err := jrpc.UpgradeToSubscription(w, r)
		w.Send(sub, err)
		if err != nil {
			return
		}
		idx := 0
		for {
			err := w.Notify(idx)
			if err != nil {
				return
			}
			idx = idx + 1
		}
	})
	if err := mux.RegisterStruct("test", new(testService)); err != nil {
		panic(err)
	}
	if err := mux.RegisterStruct("nftest", new(notificationTestService)); err != nil {
		panic(err)
	}
	return server
}

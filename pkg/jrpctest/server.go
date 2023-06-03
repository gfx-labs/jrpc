package jrpctest

import (
	"strings"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/jmux"
)

func NewServer() *jrpc.Server {
	server := jrpc.NewServer(NewRouter())
	return server
}
func NewRouter() *jmux.Mux {
	mux := jmux.NewRouter()
	//mux.HandleFunc("testservice_subscribe", func(w jrpc.ResponseWriter, r *jrpc.Request) {
	//	sub, err := jrpc.UpgradeToSubscription(w, r)
	//	w.Send(sub, err)
	//	if err != nil {
	//		return
	//	}
	//	idx := 0
	//	for {
	//		err := w.Notify(idx)
	//		if err != nil {
	//			return
	//		}
	//		idx = idx + 1
	//	}
	//})
	if err := mux.RegisterStruct("test", new(testService)); err != nil {
		panic(err)
	}
	if err := mux.RegisterStruct("nftest", new(notificationTestService)); err != nil {
		panic(err)
	}

	if err := mux.RegisterStruct("large", largeRespService{1024 * 1024 * 5 * 3}); err != nil {
		panic(err)
	}
	return mux
}
func NewRouterWithMaxSize(size int) *jmux.Mux {
	mux := jmux.NewRouter()
	//mux.HandleFunc("testservice_subscribe", func(w jrpc.ResponseWriter, r *jrpc.Request) {
	//	sub, err := jrpc.UpgradeToSubscription(w, r)
	//	w.Send(sub, err)
	//	if err != nil {
	//		return
	//	}
	//	idx := 0
	//	for {
	//		err := w.Notify(idx)
	//		if err != nil {
	//			return
	//		}
	//		idx = idx + 1
	//	}
	//})
	if err := mux.RegisterStruct("test", new(testService)); err != nil {
		panic(err)
	}
	if err := mux.RegisterStruct("nftest", new(notificationTestService)); err != nil {
		panic(err)
	}

	if err := mux.RegisterStruct("large", largeRespService{size}); err != nil {
		panic(err)
	}
	return mux
}

// largeRespService generates arbitrary-size JSON responses.
type largeRespService struct {
	length int
}

func (x largeRespService) LargeResp() string {
	return strings.Repeat("x", x.length)
}

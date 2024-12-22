package jrpctest

import (
	"strings"

	jmux2 "gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/contrib/middleware"
)

func NewRouter() *jmux2.Mux {
	mux := jmux2.NewRouter()
	mux.Use(middleware.EmptyMethodInvalid)
	mux.Use(middleware.LegacyUnderscoreReplacer)
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

	if err := mux.RegisterStruct("medium", largeRespService{1024 * 4}); err != nil {
		panic(err)
	}
	if err := mux.RegisterStruct("large", largeRespService{1024 * 1024 * 5 * 3}); err != nil {
		panic(err)
	}
	return mux
}
func NewRouterWithMaxSize(size int) *jmux2.Mux {
	mux := jmux2.NewRouter()
	mux.Use(middleware.LegacyUnderscoreReplacer)
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

package jrpctest

import (
	"encoding/json"
	"strings"

	jmux2 "github.com/gfx-labs/jrpc/contrib/jmux"
	"github.com/gfx-labs/jrpc/contrib/middleware"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
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
	mux.HandleFunc("small/largeResp", largeResp(8))
	mux.HandleFunc("medium/largeResp", largeResp(1024*4))
	mux.HandleFunc("large/largeResp", largeResp(1024*1024*5*3))
	return mux
}

func largeResp(length int) jsonrpc.HandlerFunc {
	str := json.RawMessage(`"` + strings.Repeat("x", length) + `"`)
	return func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(str, nil)
	}
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

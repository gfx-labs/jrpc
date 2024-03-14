package codecs

import (
	"context"
	"net/url"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type handlerFunc = func(bind *url.URL, h jsonrpc.Handler, opts map[string]any) error

var handlerFuncs map[string]handlerFunc = map[string]handlerFunc{}

func RegisterHandler(fn handlerFunc, names ...string) {
	for _, v := range names {
		handlerFuncs[v] = fn
	}
}

type dialerFunc = func(ctx context.Context, url string) (jsonrpc.Conn, error)

var dialers map[string]dialerFunc = map[string]dialerFunc{}

func RegisterDialer(fn dialerFunc, names ...string) {
	for _, v := range names {
		dialers[v] = fn
	}
}

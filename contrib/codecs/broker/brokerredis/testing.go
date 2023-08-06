package redis

import (
	"context"

	"gfx.cafe/open/jrpc/contrib/codecs/broker"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/server"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func ServerMaker() (*server.Server, jrpctest.ClientMaker, func()) {
	redisServer, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	connOpts := &redis.UniversalOptions{
		Addrs: []string{redisServer.Addr()},
	}
	ctx := redisServer.Ctx
	ctx, cn := context.WithCancel(ctx)
	b := CreateBroker(ctx, "jrpc", connOpts)
	s := jrpctest.NewServer()
	spokeServer := (&broker.Server{Server: s})
	go spokeServer.ServeSpoke(ctx, b)
	return s, func() codec.Conn {
			return broker.NewClient(b)
		}, func() {
			cn()
			redisServer.CtxCancel()
		}
}

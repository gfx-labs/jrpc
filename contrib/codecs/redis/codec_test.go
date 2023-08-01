package redis

import (
	"testing"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"

	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestBasicSuite(t *testing.T) {
	domain := "jrpc"

	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: func() (*server.Server, jrpctest.ClientMaker, func()) {
			redisServer := miniredis.RunT(t)
			connOpts := &redis.UniversalOptions{
				Addrs: []string{redisServer.Addr()},
			}
			ctx := redisServer.Ctx
			ss, err := CreateServerStream(ctx, domain, connOpts)
			require.NoError(t, err)
			s := jrpctest.NewServer()
			go (&Server{Server: s}).ServeRedis(ctx, ss)
			return s, func() codec.Conn {
					conn := NewClient(redis.NewUniversalClient(connOpts), domain)
					return conn
				}, func() {
					redisServer.CtxCancel()
				}
		},
	})
}

package broker

import (
	"context"
	"testing"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"

	"gfx.cafe/open/jrpc/pkg/jrpctest"
)

func TestBasicSuite(t *testing.T) {
	ctx := context.Background()
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: func() (*server.Server, jrpctest.ClientMaker, func()) {
			broker := NewChannelBroker()
			s := jrpctest.NewServer()
			spokeServer := (&Server{Server: s})
			go spokeServer.ServeSpoke(ctx, broker)
			return s, func() codec.Conn {
					conn := NewClient(broker)
					return conn
				}, func() {
				}
		},
	})
}

package websocket

import (
	"testing"

	"gfx.cafe/open/jrpc/internal/jrpctest"
)

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: ServerMaker,
	})
}

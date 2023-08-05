package websocket

import (
	"testing"

	"gfx.cafe/open/jrpc/pkg/jrpctest"
)

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: ServerMaker,
	})
}

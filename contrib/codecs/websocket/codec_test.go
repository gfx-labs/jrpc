package websocket

import (
	"testing"

	"github.com/gfx-labs/jrpc/internal/jrpctest"
)

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: ServerMaker,
	})
}

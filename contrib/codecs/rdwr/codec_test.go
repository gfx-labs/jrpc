package rdwr_test

import (
	"testing"

	"github.com/gfx-labs/jrpc/contrib/codecs/rdwr"
	"github.com/gfx-labs/jrpc/internal/jrpctest"
)

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: rdwr.ServerMaker,
	})
}

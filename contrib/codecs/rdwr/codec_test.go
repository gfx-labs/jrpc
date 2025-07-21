package rdwr_test

import (
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/internal/jrpctest"
)

func TestBasicSuite(t *testing.T) {
	jrpctest.RunBasicTestSuite(t, jrpctest.BasicTestSuiteArgs{
		ServerMaker: rdwr.ServerMaker,
	})
}

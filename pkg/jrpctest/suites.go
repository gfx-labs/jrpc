package jrpctest

import (
	"reflect"
	"testing"

	"gfx.cafe/open/jrpc"
	"github.com/stretchr/testify/require"
)

type ClientMaker func() jrpc.Conn
type ServerMaker func() (*jrpc.Server, ClientMaker, func())

type BasicTestSuiteArgs struct {
	ServerMaker ServerMaker
}

type TestContext func(t *testing.T, server *jrpc.Server, client jrpc.Conn)

func RunBasicTestSuite(t *testing.T, args BasicTestSuiteArgs) {
	var executeTest = func(t *testing.T, c TestContext) {
		server, dialer, cn := args.ServerMaker()
		defer cn()
		defer server.Stop()
		client := dialer()
		defer client.Close()
		c(t, server, client)
	}

	var makeTest = func(name string, fm func(t *testing.T, server *jrpc.Server, client jrpc.Conn)) {
		t.Run(name, func(t *testing.T) {
			executeTest(t, fm)
		})
	}
	makeTest("ClientRequest", func(t *testing.T, server *jrpc.Server, client jrpc.Conn) {
		var resp EchoResult
		err := client.Do(nil, &resp, "test_echo", []any{"hello", 10, &EchoArgs{"world"}})
		require.NoError(t, err)
		if !reflect.DeepEqual(resp, EchoResult{"hello", 10, &EchoArgs{"world"}}) {
			t.Errorf("incorrect result %#v", resp)
		}
	})
}

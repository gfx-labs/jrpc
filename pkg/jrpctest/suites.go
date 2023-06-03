package jrpctest

import (
	"context"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ClientMaker func() codec.Conn
type ServerMaker func() (*server.Server, ClientMaker, func())

type BasicTestSuiteArgs struct {
	ServerMaker ServerMaker
}

type TestContext func(t *testing.T, server *server.Server, client codec.Conn)

func RunBasicTestSuite(t *testing.T, args BasicTestSuiteArgs) {
	var executeTest = func(t *testing.T, c TestContext) {
		server, dialer, cn := args.ServerMaker()
		defer cn()
		defer server.Stop()
		client := dialer()
		defer client.Close()
		c(t, server, client)
	}

	var makeTest = func(name string, fm TestContext) {
		t.Run(name, func(t *testing.T) {
			executeTest(t, fm)
		})
	}

	t.Parallel()
	makeTest("Request", func(t *testing.T, server *server.Server, client codec.Conn) {
		var resp EchoResult
		err := client.Do(nil, &resp, "test_echo", []any{"hello", 10, &EchoArgs{"world"}})
		require.NoError(t, err)
		if !reflect.DeepEqual(resp, EchoResult{"hello", 10, &EchoArgs{"world"}}) {
			t.Errorf("incorrect result %#v", resp)
		}
	})

	makeTest("ResponseType", func(t *testing.T, server *server.Server, client codec.Conn) {
		if err := codec.CallInto(nil, client, nil, "test_echo", "hello", 10, &EchoArgs{"world"}); err != nil {
			t.Errorf("Passing nil as result should be fine, but got an error: %v", err)
		}
		var resultVar EchoResult
		// Note: passing the var, not a ref
		err := codec.CallInto(nil, client, resultVar, "test_echo", "hello", 10, &EchoArgs{"world"})
		if err == nil {
			t.Error("Passing a var as result should be an error")
		}
	})

	makeTest("BatchRequest", func(t *testing.T, server *server.Server, client codec.Conn) {
		batch := []*codec.BatchElem{
			{
				Method: "test_echo",
				Params: []any{"hello", 10, &EchoArgs{"world"}},
				Result: new(EchoResult),
			},
			{
				Method: "test_echo",
				Params: []any{"hello2", 11, &EchoArgs{"world"}},
				Result: new(EchoResult),
			},
			{
				Method: "no_such_method",
				Params: []any{1, 2, 3},
				Result: new(int),
			},
		}
		if err := client.BatchCall(nil, batch...); err != nil {
			t.Fatal(err)
		}
		wantResult := []*codec.BatchElem{
			{
				Method: "test_echo",
				Params: []any{"hello", 10, &EchoArgs{"world"}},
				Result: &EchoResult{"hello", 10, &EchoArgs{"world"}},
			},
			{
				Method: "test_echo",
				Params: []any{"hello2", 11, &EchoArgs{"world"}},
				Result: &EchoResult{"hello2", 11, &EchoArgs{"world"}},
			},
			{
				Method: "no_such_method",
				Params: []any{1, 2, 3},
				Result: new(int),
				Error:  &codec.JsonError{Code: -32601, Message: "the method no_such_method does not exist/is not available"},
			},
		}
		require.EqualValues(t, len(batch), len(wantResult))
		for i := range batch {
			a := batch[i]
			b := batch[i]
			assert.EqualValuesf(t, a.Method, b.Method, "item %d", i)
			assert.EqualValuesf(t, a.Result, b.Result, "item %d", i)
			assert.EqualValuesf(t, a.Params, b.Params, "item %d", i)
			if a.Error != nil {
				assert.EqualValuesf(t, a.Error, b.Error, "item %d", i)
			}
		}
	})

	makeTest("ResposeType", func(t *testing.T, server *server.Server, client codec.Conn) {
		if err := codec.CallInto(nil, client, nil, "test_echo", "hello", 10, &EchoArgs{"world"}); err != nil {
			t.Errorf("Passing nil as result should be fine, but got an error: %v", err)
		}
		var resultVar EchoResult
		// Note: passing the var, not a ref
		err := codec.CallInto(nil, client, resultVar, "test_echo", "hello", 10, &EchoArgs{"world"})
		if err == nil {
			t.Error("Passing a var as result should be an error")
		}
	})

	makeTest("ErrorReturnType", func(t *testing.T, server *server.Server, client codec.Conn) {
		var resp any
		err := codec.CallInto(nil, client, &resp, "test_returnError")
		require.Error(t, err)

		// Check code.
		if e, ok := err.(codec.Error); !ok {
			t.Fatalf("client did not return rpc.Error, got %#v", e)
		} else if e.ErrorCode() != (testError{}.ErrorCode()) {
			t.Fatalf("wrong error code %d, want %d", e.ErrorCode(), testError{}.ErrorCode())
		}
		// Check data.
		if e, ok := err.(codec.DataError); !ok {
			t.Fatalf("client did not return rpc.DataError, got %#v", e)
		} else if e.ErrorData() != (testError{}.ErrorData()) {
			t.Fatalf("wrong error data %#v, want %#v", e.ErrorData(), testError{}.ErrorData())
		}
	})
	makeTest("Notify", func(t *testing.T, server *server.Server, client codec.Conn) {
		if err := client.Notify(context.Background(), "test_echo", []any{"hello", 10, &EchoArgs{"world"}}); err != nil {
			t.Fatal(err)
		}
	})

	makeTest("context cancel", func(t *testing.T, server *server.Server, client codec.Conn) {
		maxContextCancelTimeout := 300 * time.Millisecond
		// The actual test starts here.
		var (
			wg       sync.WaitGroup
			nreqs    = 10
			ncallers = 10
		)
		caller := func(index int) {
			defer wg.Done()
			for i := 0; i < nreqs; i++ {
				var (
					ctx     context.Context
					cancel  func()
					timeout = time.Duration(rand.Int63n(int64(maxContextCancelTimeout)))
				)
				if index < ncallers/2 {
					// For half of the callers, create a context without deadline
					// and cancel it later.
					ctx, cancel = context.WithCancel(context.Background())
					time.AfterFunc(timeout, cancel)
				} else {
					// For the other half, create a context with a deadline instead. This is
					// different because the context deadline is used to set the socket write
					// deadline.
					ctx, cancel = context.WithTimeout(context.Background(), timeout)
				}

				// Now perform a call with the context.
				// The key thing here is that no call will ever complete successfully.
				err := codec.CallInto(ctx, client, nil, "test_block")
				switch {
				case err == nil:
					_, hasDeadline := ctx.Deadline()
					t.Errorf("no error for call with %v wait time (deadline: %v)", timeout, hasDeadline)
					// default:
					// 	t.Logf("got expected error with %v wait time: %v", timeout, err)
				}
				cancel()
			}
		}
		wg.Add(ncallers)
		for i := 0; i < ncallers; i++ {
			go caller(i)
		}
		wg.Wait()
	})
}

type ServerRemaker func(address string) (*server.Server, ClientMaker, func())

type ReconnectTestSuiteArgs struct {
	ServerMaker ServerMaker
}

func RunReconnectSuite(t *testing.T, args BasicTestSuiteArgs) {
	server, dialer, cn := args.ServerMaker()
	defer cn()
	defer server.Stop()
	client := dialer()
	defer client.Close()

}

// This test checks that requests made through Call can be canceled by canceling
// the context.
func cancelTester(t *testing.T, server *server.Server, client codec.Conn) {
	maxContextCancelTimeout := 300 * time.Millisecond

	// The actual test starts here.
	var (
		wg       sync.WaitGroup
		nreqs    = 10
		ncallers = 10
	)
	caller := func(index int) {
		defer wg.Done()
		for i := 0; i < nreqs; i++ {
			var (
				ctx     context.Context
				cancel  func()
				timeout = time.Duration(rand.Int63n(int64(maxContextCancelTimeout)))
			)
			if index < ncallers/2 {
				// For half of the callers, create a context without deadline
				// and cancel it later.
				ctx, cancel = context.WithCancel(context.Background())
				time.AfterFunc(timeout, cancel)
			} else {
				// For the other half, create a context with a deadline instead. This is
				// different because the context deadline is used to set the socket write
				// deadline.
				ctx, cancel = context.WithTimeout(context.Background(), timeout)
			}

			// Now perform a call with the context.
			// The key thing here is that no call will ever complete successfully.
			err := codec.CallInto(ctx, client, nil, "test_block")
			switch {
			case err == nil:
				_, hasDeadline := ctx.Deadline()
				t.Errorf("no error for call with %v wait time (deadline: %v)", timeout, hasDeadline)
				// default:
				// 	t.Logf("got expected error with %v wait time: %v", timeout, err)
			}
			cancel()
		}
	}
	wg.Add(ncallers)
	for i := 0; i < ncallers; i++ {
		go caller(i)
	}
	wg.Wait()
}

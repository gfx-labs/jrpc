package jrpctest

import (
	"context"
	"embed"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ClientMaker func() jsonrpc.Conn
type ServerMaker func() (jsonrpc.Handler, ClientMaker, func())

type BasicTestSuiteArgs struct {
	ServerMaker ServerMaker
}

type TestContext func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn)
type BenchContext func(t *testing.B, server jsonrpc.Handler, client jsonrpc.Conn)

func TestExecutor(sm ServerMaker) func(t *testing.T, c TestContext) {
	return func(t *testing.T, c TestContext) {
		server, dialer, cn := sm()
		defer cn()
		client := dialer()
		defer client.Close()
		c(t, server, client)
	}
}
func BenchExecutor(sm ServerMaker) func(t *testing.B, c BenchContext) {
	return func(t *testing.B, c BenchContext) {
		server, dialer, cn := sm()
		defer cn()
		client := dialer()
		defer client.Close()
		c(t, server, client)
	}
}

// go:embed testdata/
var testData embed.FS

func RunBasicTestSuite(t *testing.T, args BasicTestSuiteArgs) {
	var executeTest = TestExecutor(args.ServerMaker)

	var makeTest = func(name string, fm TestContext) {
		t.Run(name, func(t *testing.T) {
			executeTest(t, fm)
		})
	}

	t.Parallel()
	makeTest("Request", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
		var resp EchoResult
		err := client.Do(nil, &resp, "test_echo", []any{"hello", 10, &EchoArgs{"world"}})
		require.NoError(t, err)
		if !reflect.DeepEqual(resp, EchoResult{"hello", 10, &EchoArgs{"world"}}) {
			t.Errorf("incorrect result %#v", resp)
		}
	})

	makeTest("ResponseType", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
		err := jsonrpc.CallInto(nil, client, nil, "test_echo", "hello", 10, &EchoArgs{"world"})
		assert.NoErrorf(t, err, "passing nil as result should be ok")
		var resultVar EchoResult
		// Note: passing the var, not a ref
		err = jsonrpc.CallInto(nil, client, resultVar, "test_echo", "hello", 10, &EchoArgs{"world"})
		assert.Error(t, err, "passing var as nil gives error")
	})

	makeTest("ResposeType2", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
		if err := jsonrpc.CallInto(nil, client, nil, "test_echo", "hello", 10, &EchoArgs{"world"}); err != nil {
			t.Errorf("Passing nil as result should be fine, but got an error: %v", err)
		}
		var resultVar EchoResult
		// Note: passing the var, not a ref
		err := jsonrpc.CallInto(nil, client, resultVar, "test_echo", "hello", 10, &EchoArgs{"world"})
		if err == nil {
			t.Error("Passing a var as result should be an error")
		}
	})

	makeTest("ErrorReturnType", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
		var resp any
		err := jsonrpc.CallInto(nil, client, &resp, "test_returnError")
		require.Error(t, err)

		// Check code.
		if e, ok := err.(jsonrpc.Error); !ok {
			t.Fatalf("client did not return rpc.Error, got %#v", e)
		} else if e.ErrorCode() != (testError{}.ErrorCode()) {
			t.Fatalf("wrong error code %d, want %d", e.ErrorCode(), testError{}.ErrorCode())
		}
		// Check data.
		if e, ok := err.(jsonrpc.DataError); !ok {
			t.Fatalf("client did not return rpc.DataError, got %#v", e)
		} else if e.ErrorData() != (testError{}.ErrorData()) {
			t.Fatalf("wrong error data %#v, want %#v", e.ErrorData(), testError{}.ErrorData())
		}
	})
	makeTest("Notify", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
		if c, ok := client.(jsonrpc.Conn); ok {
			if err := c.Notify(context.Background(), "test_echo", []any{"hello", 10, &EchoArgs{"world"}}); err != nil {
				t.Fatal(err)
			}
		}
	})

	makeTest("context cancel", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
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
				err := jsonrpc.CallInto(ctx, client, nil, "test_block")
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

	makeTest("big", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
		var (
			wg       sync.WaitGroup
			nreqs    = 2
			ncallers = 10
		)
		wg.Add(ncallers)
		// create a bunch of parallel requests with lots of data to see if any buffers are overwritten causing a failure
		for i := 0; i < ncallers; i++ {
			go func() {
				defer wg.Done()

				for j := 0; j < nreqs; j++ {
					if err := jsonrpc.CallInto(context.Background(), client, nil, "large_largeResp"); err != nil {
						t.Error(err)
						return
					}
				}
			}()
		}

		wg.Wait()
	})

	makeTest("", func(t *testing.T, server jsonrpc.Handler, client jsonrpc.Conn) {
	})
}

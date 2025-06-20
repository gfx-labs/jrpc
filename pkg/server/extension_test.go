package server_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"gfx.cafe/open/jrpc/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseWriterExtension(t *testing.T) {
	t.Run("adds extension fields to response", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readbuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			// Add extensions directly
			err := w.Extension("metadata", map[string]string{"version": "1.0"})
			require.NoError(t, err)

			err = w.Extension("timestamp", 1234567890)
			require.NoError(t, err)

			// Send response
			err = w.Send("hello world", nil)
			require.NoError(t, err)
		})

		// Process request
		ctx := context.Background()
		go srv.ServeCodec(ctx, codec, handler)

		// Create and send request
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "test",
		}
		reqBytes, _ := json.Marshal(req)

		rd.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := rd.Write(reqBytes)
		require.NoError(t, err)

		// Read response
		rd.SetReadDeadline(time.Now().Add(5 * time.Second))
		responseBytes := make([]byte, 1024)
		n, err := readbuf.Read(responseBytes)
		require.NoError(t, err)
		responseBytes = responseBytes[:n]

		// Parse using JSON-RPC decoder
		msgs, isBatch := jsonrpc.ParseMessage(responseBytes)
		require.False(t, isBatch)
		require.Len(t, msgs, 1)

		msg := msgs[0]
		require.NotNil(t, msg)
		require.NotNil(t, msg.ID)
		require.Nil(t, msg.Error)
		require.NotNil(t, msg.Result)

		// Verify extensions
		require.NotNil(t, msg.Extensions)
		assert.Contains(t, msg.Extensions, "metadata")
		assert.Contains(t, msg.Extensions, "timestamp")

		// Verify extension values
		var metadata map[string]string
		err = json.Unmarshal(msg.Extensions["metadata"], &metadata)
		require.NoError(t, err)
		assert.Equal(t, "1.0", metadata["version"])

		var timestamp int
		err = json.Unmarshal(msg.Extensions["timestamp"], &timestamp)
		require.NoError(t, err)
		assert.Equal(t, 1234567890, timestamp)

		// Verify result
		resultBytes, err := io.ReadAll(msg.Result)
		require.NoError(t, err)
		var result string
		err = json.Unmarshal(resultBytes, &result)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("extensions appear before result in response", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readBuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			w.Extension("custom", "value")
			w.Send("result", nil)
		})

		// Process request
		ctx := context.Background()
		go srv.ServeCodec(ctx, codec, handler)

		// Create and send request
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "test",
		}
		reqBytes, _ := json.Marshal(req)

		rd.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := rd.Write(reqBytes)
		require.NoError(t, err)

		// Read raw response to check field order
		rd.SetReadDeadline(time.Now().Add(5 * time.Second))
		responseBytes := make([]byte, 1024)
		n, err := readBuf.Read(responseBytes)
		require.NoError(t, err)
		responseBytes = responseBytes[:n]

		// Find positions of fields
		customPos := bytes.Index(responseBytes, []byte(`"custom"`))
		resultPos := bytes.Index(responseBytes, []byte(`"result"`))

		assert.Greater(t, customPos, 0, "custom field should be present")
		assert.Greater(t, resultPos, 0, "result field should be present")
		assert.Less(t, customPos, resultPos, "custom field should appear before result field")
	})

	t.Run("extensions work with error responses", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readBuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			w.Extension("errorCode", "ERR_001")
			w.Send(nil, jsonrpc.NewInvalidRequestError("test error"))
		})

		// Process request
		ctx := context.Background()
		go srv.ServeCodec(ctx, codec, handler)

		// Create and send request
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "test",
		}
		reqBytes, _ := json.Marshal(req)

		rd.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := rd.Write(reqBytes)
		require.NoError(t, err)

		// Read response
		rd.SetReadDeadline(time.Now().Add(5 * time.Second))
		responseBytes := make([]byte, 1024)
		n, err := readBuf.Read(responseBytes)
		require.NoError(t, err)
		responseBytes = responseBytes[:n]

		// Parse using JSON-RPC decoder
		msgs, isBatch := jsonrpc.ParseMessage(responseBytes)
		require.False(t, isBatch)
		require.Len(t, msgs, 1)

		msg := msgs[0]
		require.NotNil(t, msg)
		require.NotNil(t, msg.ID)
		require.NotNil(t, msg.Error)
		require.Nil(t, msg.Result)

		// Verify extension
		require.NotNil(t, msg.Extensions)
		assert.Contains(t, msg.Extensions, "errorCode")

		// Verify extension value
		var errorCode string
		err = json.Unmarshal(msg.Extensions["errorCode"], &errorCode)
		require.NoError(t, err)
		assert.Equal(t, "ERR_001", errorCode)
	})

	t.Run("cannot add extensions after Send", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			// Send response first
			err := w.Send("result", nil)
			require.NoError(t, err)

			// Try to add extension after Send - should fail
			err = w.Extension("late", "extension")
			assert.Equal(t, jsonrpc.ErrSendAlreadyCalled, err)
		})

		// Process request
		ctx := context.Background()
		go srv.ServeCodec(ctx, codec, handler)

		// Create and send request
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "test",
		}
		reqBytes, _ := json.Marshal(req)

		rd.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := rd.Write(reqBytes)
		require.NoError(t, err)

		// Give time for processing
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("handles marshaling errors", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readBuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		// Create a type that cannot be marshaled to JSON
		type unmarshalable struct {
			Ch chan int
		}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			// Try to add an unmarshalable value
			err := w.Extension("bad", unmarshalable{Ch: make(chan int)})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "json")

			// Can still send response
			err = w.Send("ok", nil)
			require.NoError(t, err)
		})

		// Process request
		ctx := context.Background()
		go srv.ServeCodec(ctx, codec, handler)

		// Create and send request
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "test",
		}
		reqBytes, _ := json.Marshal(req)

		rd.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := rd.Write(reqBytes)
		require.NoError(t, err)

		// Read response
		rd.SetReadDeadline(time.Now().Add(5 * time.Second))
		responseBytes := make([]byte, 1024)
		n, err := readBuf.Read(responseBytes)
		require.NoError(t, err)
		responseBytes = responseBytes[:n]

		// Parse using JSON-RPC decoder
		msgs, isBatch := jsonrpc.ParseMessage(responseBytes)
		require.False(t, isBatch)
		require.Len(t, msgs, 1)

		msg := msgs[0]
		require.NotNil(t, msg)
		require.NotNil(t, msg.Result)

		// Should have result but no "bad" extension
		if msg.Extensions != nil {
			assert.NotContains(t, msg.Extensions, "bad")
		}
	})
}

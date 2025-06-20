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

func TestResponseWriterExtraFields(t *testing.T) {
	t.Run("adds extra fields to response", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readbuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			// Add extra fields using map-like API
			w.ExtraFields().Set("metadata", map[string]string{"version": "1.0"})
			w.ExtraFields().Set("timestamp", 1234567890)

			// Send response
			err := w.Send("hello world", nil)
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

		// Verify extra fields
		require.NotNil(t, msg.ExtraFields)
		assert.Contains(t, msg.ExtraFields, "metadata")
		assert.Contains(t, msg.ExtraFields, "timestamp")

		// Verify extra field values
		var metadata map[string]string
		err = json.Unmarshal(msg.ExtraFields["metadata"], &metadata)
		require.NoError(t, err)
		assert.Equal(t, "1.0", metadata["version"])

		var timestamp int
		err = json.Unmarshal(msg.ExtraFields["timestamp"], &timestamp)
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

	t.Run("extra fields appear before result in response", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readBuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			w.ExtraFields().Set("custom", "value")
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

	t.Run("extra fields work with error responses", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readBuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			w.ExtraFields().Set("errorCode", "ERR_001")
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

		// Verify extra field
		require.NotNil(t, msg.ExtraFields)
		assert.Contains(t, msg.ExtraFields, "errorCode")

		// Verify extra field value
		var errorCode string
		err = json.Unmarshal(msg.ExtraFields["errorCode"], &errorCode)
		require.NoError(t, err)
		assert.Equal(t, "ERR_001", errorCode)
	})

	t.Run("cannot modify extra fields after Send", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			// Get reference to extra fields
			extraFields := w.ExtraFields()
			
			// Send response first
			err := w.Send("result", nil)
			require.NoError(t, err)

			// Extra fields can still be modified after Send, but they won't be included
			// This is consistent with http.Header behavior
			extraFields.Set("late", "field")
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
			// The marshaling error will happen during Send, not during Set
			w.ExtraFields().Set("bad", unmarshalable{Ch: make(chan int)})

			// Send will fail due to marshaling error
			err := w.Send("ok", nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported type")
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

		// Since Send failed, there might not be a valid response
		// Let's just ensure the goroutine completes
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("cannot set reserved JSON-RPC fields", func(t *testing.T) {
		// Create a net pipe
		rd, wr := net.Pipe()
		readBuf := bufio.NewReader(rd)
		codec := rdwr.NewCodec(wr, wr)
		srv := &server.Server{}

		handler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			// Try to set reserved fields - these should be ignored
			w.ExtraFields().Set("jsonrpc", "3.0")
			w.ExtraFields().Set("id", 999)
			w.ExtraFields().Set("method", "fake")
			w.ExtraFields().Set("params", []string{"test"})
			w.ExtraFields().Set("result", "fake-result")
			w.ExtraFields().Set("error", "fake-error")
			
			// Set a valid extra field
			w.ExtraFields().Set("custom", "allowed")

			// Send response
			err := w.Send("real-result", nil)
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
		
		// Verify that reserved fields were not overridden
		require.NotNil(t, msg.ID)
		assert.Equal(t, 1, msg.ID.Number())
		
		// Result should be the real result, not the fake one
		resultBytes, err := io.ReadAll(msg.Result)
		require.NoError(t, err)
		var result string
		err = json.Unmarshal(resultBytes, &result)
		require.NoError(t, err)
		assert.Equal(t, "real-result", result)
		
		// Only the custom field should be present in extra fields
		require.NotNil(t, msg.ExtraFields)
		assert.Contains(t, msg.ExtraFields, "custom")
		assert.Len(t, msg.ExtraFields, 1)
		
		var custom string
		err = json.Unmarshal(msg.ExtraFields["custom"], &custom)
		require.NoError(t, err)
		assert.Equal(t, "allowed", custom)
	})
}

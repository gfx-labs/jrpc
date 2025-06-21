package jsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNull(t *testing.T) {
	null := NewNull()
	assert.Equal(t, json.RawMessage("null"), null)
	assert.Equal(t, NullString, string(null))
}

func TestNewStringReader(t *testing.T) {
	content := "test content"
	reader := NewStringReader(content)

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	// Should be closeable
	err = reader.Close()
	assert.NoError(t, err)
}

func TestMessageMarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		want    map[string]any
		wantErr bool
	}{
		{
			name: "request with ID and method",
			msg: Message{
				ID:     NewStringIDPtr("123"),
				Method: "test_method",
				Params: json.RawMessage(`{"param": "value"}`),
			},
			want: map[string]any{
				"jsonrpc": "2.0",
				"id":      "123",
				"method":  "test_method",
				"params":  map[string]any{"param": "value"},
			},
		},
		{
			name: "notification without ID",
			msg: Message{
				Method: "notify",
				Params: json.RawMessage(`[1, 2, 3]`),
			},
			want: map[string]any{
				"jsonrpc": "2.0",
				"method":  "notify",
				"params":  []any{float64(1), float64(2), float64(3)},
			},
		},
		// Skip testing Result field in MarshalJSON as it requires proper streaming setup
		{
			name: "error response",
			msg: Message{
				ID: NewStringIDPtr("789"),
				Error: &JsonError{
					Code:    -32600,
					Message: "Invalid Request",
					Data:    "Additional info",
				},
			},
			want: map[string]any{
				"jsonrpc": "2.0",
				"id":      "789",
				"error": map[string]any{
					"code":    float64(-32600),
					"message": "Invalid Request",
					"data":    "Additional info",
				},
			},
		},
		{
			name: "message with extra fields",
			msg: Message{
				ID:     NewStringIDPtr("extra"),
				Method: "test",
				ExtraFields: map[string]json.RawMessage{
					"custom1": json.RawMessage(`"value1"`),
					"custom2": json.RawMessage(`{"nested": "object"}`),
				},
			},
			want: map[string]any{
				"jsonrpc": "2.0",
				"id":      "extra",
				"method":  "test",
				"custom1": "value1",
				"custom2": map[string]any{"nested": "object"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.msg.MarshalJSON()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			var got map[string]any
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMessageUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Message
		wantErr bool
		errMsg  string
	}{
		{
			name:  "valid request",
			input: `{"jsonrpc":"2.0","id":"123","method":"test","params":{"key":"value"}}`,
			want: Message{
				ID:     NewStringIDPtr("123"),
				Method: "test",
				Params: json.RawMessage(`{"key":"value"}`),
			},
		},
		{
			name:  "valid notification",
			input: `{"jsonrpc":"2.0","method":"notify","params":[1,2,3]}`,
			want: Message{
				Method: "notify",
				Params: json.RawMessage(`[1,2,3]`),
			},
		},
		{
			name:  "valid response",
			input: `{"jsonrpc":"2.0","id":456,"result":{"success":true}}`,
			want: Message{
				ID: NewNumberIDPtr(456),
			},
		},
		{
			name:  "valid error response",
			input: `{"jsonrpc":"2.0","id":"error","error":{"code":-32600,"message":"Invalid Request"}}`,
			want: Message{
				ID: NewStringIDPtr("error"),
				Error: &JsonError{
					Code:    -32600,
					Message: "Invalid Request",
				},
			},
		},
		{
			name:  "message with extra fields",
			input: `{"jsonrpc":"2.0","id":"extra","method":"test","custom":"field","another":123}`,
			want: Message{
				ID:     NewStringIDPtr("extra"),
				Method: "test",
				ExtraFields: map[string]json.RawMessage{
					"custom":  json.RawMessage(`"field"`),
					"another": json.RawMessage(`123`),
				},
			},
		},
		{
			name:    "invalid version",
			input:   `{"jsonrpc":"1.0","method":"test"}`,
			wantErr: true,
			errMsg:  "Invalid Version",
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:  "null id",
			input: `{"jsonrpc":"2.0","id":null,"method":"test"}`,
			want: Message{
				ID:     NewNullIDPtr(),
				Method: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg Message
			err := msg.UnmarshalJSON([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)

			// Compare messages
			assert.Equal(t, tt.want.Method, msg.Method)

			// Compare IDs
			if tt.want.ID != nil {
				if msg.ID == nil {
					assert.Fail(t, "Expected ID but got nil")
				} else {
					wantJSON, _ := tt.want.ID.MarshalJSON()
					gotJSON, _ := msg.ID.MarshalJSON()
					assert.JSONEq(t, string(wantJSON), string(gotJSON))
				}
			} else {
				assert.Nil(t, msg.ID)
			}

			// Compare params
			if tt.want.Params != nil {
				assert.JSONEq(t, string(tt.want.Params), string(msg.Params))
			}

			// Compare errors
			if tt.want.Error != nil {
				jsonErr, ok := msg.Error.(*JsonError)
				require.True(t, ok)
				wantErr, ok := tt.want.Error.(*JsonError)
				require.True(t, ok)
				assert.Equal(t, wantErr.Code, jsonErr.Code)
				assert.Equal(t, wantErr.Message, jsonErr.Message)
			}

			// Compare extra fields
			if tt.want.ExtraFields != nil {
				assert.Equal(t, len(tt.want.ExtraFields), len(msg.ExtraFields))
				for k, v := range tt.want.ExtraFields {
					assert.JSONEq(t, string(v), string(msg.ExtraFields[k]))
				}
			}

			// Check result separately as it's an io.ReadCloser
			if tt.input == `{"jsonrpc":"2.0","id":456,"result":{"success":true}}` {
				assert.NotNil(t, msg.Result)
				data, err := io.ReadAll(msg.Result)
				require.NoError(t, err)
				assert.JSONEq(t, `{"success":true}`, string(data))
			}
		})
	}
}

func TestMarshalMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     *Message
		wantErr bool
	}{
		{
			name: "marshal request",
			msg: &Message{
				ID:     NewStringIDPtr("test"),
				Method: "testMethod",
				Params: json.RawMessage(`{"param": "value"}`),
			},
		},
		{
			name: "marshal error response",
			msg: &Message{
				ID: NewStringIDPtr("error"),
				Error: &JsonError{
					Code:    -32700,
					Message: "Parse error",
				},
			},
		},
		{
			name: "marshal with extra fields",
			msg: &Message{
				ID:     NewStringIDPtr("extra"),
				Method: "test",
				ExtraFields: map[string]json.RawMessage{
					"custom": json.RawMessage(`"value"`),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			enc := jx.NewStreamingEncoder(buf, 1024)

			err := MarshalMessage(tt.msg, enc)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			err = enc.Close()
			require.NoError(t, err)

			// Verify it's valid JSON
			var result map[string]any
			err = json.Unmarshal(buf.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, "2.0", result["jsonrpc"])
		})
	}
}

func TestUnmarshalMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "valid request",
			input: `{"jsonrpc":"2.0","id":"1","method":"test"}`,
		},
		{
			name:  "valid response",
			input: `{"jsonrpc":"2.0","id":"1","result":"success"}`,
		},
		{
			name:  "valid error",
			input: `{"jsonrpc":"2.0","id":"1","error":{"code":-32600,"message":"Invalid Request"}}`,
		},
		{
			name:    "invalid json",
			input:   `{invalid`,
			wantErr: true,
		},
		{
			name:    "wrong version",
			input:   `{"jsonrpc":"1.0","method":"test"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := jx.DecodeBytes([]byte(tt.input))
			msg := &Message{}

			err := UnmarshalMessage(msg, dec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, msg)
		})
	}
}

func TestJsonError(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		err := &JsonError{
			Code:    -32600,
			Message: "Invalid Request",
		}
		assert.Equal(t, "Invalid Request", err.Error())

		// Test with empty message
		err2 := &JsonError{
			Code: -32700,
		}
		assert.Equal(t, "json-rpc error -32700", err2.Error())
	})

	t.Run("ErrorCode method", func(t *testing.T) {
		err := &JsonError{Code: -32600}
		assert.Equal(t, -32600, err.ErrorCode())
	})

	t.Run("ErrorData method", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		err := &JsonError{
			Code: -32600,
			Data: data,
		}
		assert.Equal(t, data, err.ErrorData())
	})
}

func TestIsBatchMessage(t *testing.T) {
	tests := []struct {
		name  string
		input json.RawMessage
		want  bool
	}{
		{
			name:  "batch message",
			input: json.RawMessage(`[{"jsonrpc":"2.0","method":"test"}]`),
			want:  true,
		},
		{
			name: "batch with whitespace",
			input: json.RawMessage(`
[{"jsonrpc":"2.0"}]`),
			want: true,
		},
		{
			name:  "single message",
			input: json.RawMessage(`{"jsonrpc":"2.0","method":"test"}`),
			want:  false,
		},
		{
			name:  "single with whitespace",
			input: json.RawMessage(`   {"jsonrpc":"2.0"}`),
			want:  false,
		},
		{
			name:  "empty",
			input: json.RawMessage(``),
			want:  false,
		},
		{
			name: "only whitespace",
			input: json.RawMessage(`
	`),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBatchMessage(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseMessage(t *testing.T) {
	t.Run("single message", func(t *testing.T) {
		input := json.RawMessage(`{"jsonrpc":"2.0","id":"1","method":"test"}`)
		msgs, isBatch := ParseMessage(input)

		assert.False(t, isBatch)
		assert.Len(t, msgs, 1)
		assert.Equal(t, "test", msgs[0].Method)
		idJSON, _ := msgs[0].ID.MarshalJSON()
		assert.Equal(t, `"1"`, string(idJSON))
	})

	t.Run("batch messages", func(t *testing.T) {
		input := json.RawMessage(`[
			{"jsonrpc":"2.0","id":"1","method":"test1"},
			{"jsonrpc":"2.0","id":"2","method":"test2"}
		]`)
		msgs, isBatch := ParseMessage(input)

		assert.True(t, isBatch)
		assert.Len(t, msgs, 2)
		assert.Equal(t, "test1", msgs[0].Method)
		assert.Equal(t, "test2", msgs[1].Method)
	})

	t.Run("batch with message missing version", func(t *testing.T) {
		// Note: Messages without jsonrpc field are still parsed, they just go into ExtraFields
		input := json.RawMessage(`[
			{"jsonrpc":"2.0","id":"1","method":"test1"},
			{"invalid":"message"},
			{"jsonrpc":"2.0","id":"3","method":"test3"}
		]`)
		msgs, isBatch := ParseMessage(input)

		assert.True(t, isBatch)
		assert.Len(t, msgs, 3)
		assert.Equal(t, "test1", msgs[0].Method)
		// Message without jsonrpc field is still parsed
		assert.NotNil(t, msgs[1])
		assert.Equal(t, "", msgs[1].Method)
		assert.Contains(t, msgs[1].ExtraFields, "invalid")
		assert.Equal(t, "test3", msgs[2].Method)
	})

	t.Run("invalid json", func(t *testing.T) {
		input := json.RawMessage(`{invalid}`)
		msgs, isBatch := ParseMessage(input)

		assert.False(t, isBatch)
		assert.Len(t, msgs, 1)
		// Invalid message results in empty Message struct
		assert.Equal(t, "", msgs[0].Method)
		assert.Nil(t, msgs[0].ID)
	})
}

func TestReadMessage(t *testing.T) {
	t.Run("object message", func(t *testing.T) {
		input := `{"jsonrpc":"2.0","id":"1","method":"test"}`
		dec := jx.DecodeBytes([]byte(input))

		msgs, isBatch := ReadMessage(dec)
		assert.False(t, isBatch)
		assert.Len(t, msgs, 1)
		assert.Equal(t, "test", msgs[0].Method)
	})

	t.Run("array message", func(t *testing.T) {
		input := `[{"jsonrpc":"2.0","id":"1","method":"test"}]`
		dec := jx.DecodeBytes([]byte(input))

		msgs, isBatch := ReadMessage(dec)
		assert.True(t, isBatch)
		assert.Len(t, msgs, 1)
		assert.Equal(t, "test", msgs[0].Method)
	})

	t.Run("neither object nor array", func(t *testing.T) {
		input := `"string"`
		dec := jx.DecodeBytes([]byte(input))

		msgs, isBatch := ReadMessage(dec)
		assert.False(t, isBatch)
		assert.Len(t, msgs, 1)
		assert.Equal(t, &Message{}, msgs[0])
	})
}

func TestMessageString(t *testing.T) {
	msg := Message{
		ID:     NewStringIDPtr("test"),
		Method: "testMethod",
	}

	str := msg.String()
	assert.Contains(t, str, `"jsonrpc":"2.0"`)
	assert.Contains(t, str, `"id":"test"`)
	assert.Contains(t, str, `"method":"testMethod"`)
}

func TestMarshalError(t *testing.T) {
	// Test various error scenarios
	t.Run("marshal json error", func(t *testing.T) {
		err := &JsonError{
			Code:    -32700,
			Message: "Parse error",
			Data:    "Additional info",
		}

		data := MarshalError(err)

		var result map[string]any
		unmarshalErr := json.Unmarshal(data, &result)
		require.NoError(t, unmarshalErr)

		assert.Equal(t, float64(-32700), result["code"])
		assert.Equal(t, "Parse error", result["message"])
		assert.Equal(t, "Additional info", result["data"])
	})

	t.Run("marshal generic error", func(t *testing.T) {
		err := errors.New("generic error")
		data := MarshalError(err)

		var result map[string]any
		unmarshalErr := json.Unmarshal(data, &result)
		require.NoError(t, unmarshalErr)

		assert.Equal(t, float64(-32000), result["code"])
		assert.Equal(t, "generic error", result["message"])
	})
}

// Benchmarks
func BenchmarkMessageMarshalJSON(b *testing.B) {
	msg := Message{
		ID:     NewStringIDPtr("bench"),
		Method: "benchmark",
		Params: json.RawMessage(`{"key":"value","number":123}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msg.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageUnmarshalJSON(b *testing.B) {
	data := []byte(`{"jsonrpc":"2.0","id":"bench","method":"benchmark","params":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg Message
		err := msg.UnmarshalJSON(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIsBatchMessage(b *testing.B) {
	single := json.RawMessage(`{"jsonrpc":"2.0","method":"test"}`)
	batch := json.RawMessage(`[
		{"jsonrpc":"2.0","id":"1","method":"test1"},
		{"jsonrpc":"2.0","id":"2","method":"test2"},
		{"jsonrpc":"2.0","id":"3","method":"test3"}
	]`)

	b.Run("single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsBatchMessage(single)
		}
	})

	b.Run("batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsBatchMessage(batch)
		}
	})
}

func BenchmarkParseMessage(b *testing.B) {
	single := json.RawMessage(`{"jsonrpc":"2.0","id":"1","method":"test","params":{"key":"value"}}`)
	batch := json.RawMessage(`[
		{"jsonrpc":"2.0","id":"1","method":"test1"},
		{"jsonrpc":"2.0","id":"2","method":"test2"},
		{"jsonrpc":"2.0","id":"3","method":"test3"}
	]`)

	b.Run("single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ParseMessage(single)
		}
	})

	b.Run("batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ParseMessage(batch)
		}
	})
}


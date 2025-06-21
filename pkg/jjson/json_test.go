package jjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs
type testStruct struct {
	Name    string                 `json:"name"`
	Age     int                    `json:"age"`
	Active  bool                   `json:"active"`
	Tags    []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

type customMarshaler struct {
	Value string
}

func (c customMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"custom": c.Value})
}

type customUnmarshaler struct {
	Value string
}

func (c *customUnmarshaler) UnmarshalJSON(data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	c.Value = m["custom"]
	return nil
}

type errorMarshaler struct{}

func (e errorMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal error")
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "simple string",
			input: "hello world",
			want:  `"hello world"`,
		},
		{
			name:  "number",
			input: 42,
			want:  "42",
		},
		{
			name:  "boolean",
			input: true,
			want:  "true",
		},
		{
			name:  "null",
			input: nil,
			want:  "null",
		},
		{
			name: "struct",
			input: testStruct{
				Name:   "John",
				Age:    30,
				Active: true,
				Tags:   []string{"go", "json"},
				Metadata: map[string]interface{}{
					"level": "advanced",
					"score": 9.5,
				},
			},
			want: `{"name":"John","age":30,"active":true,"tags":["go","json"],"metadata":{"level":"advanced","score":9.5}}`,
		},
		{
			name:  "custom marshaler",
			input: customMarshaler{Value: "test"},
			want:  `{"custom":"test"}`,
		},
		{
			name:    "error marshaler",
			input:   errorMarshaler{},
			wantErr: true,
		},
		{
			name:  "empty slice",
			input: []int{},
			want:  "[]",
		},
		{
			name:  "empty map",
			input: map[string]int{},
			want:  "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:   "simple string",
			input:  `"hello world"`,
			target: new(string),
			want:   "hello world",
		},
		{
			name:   "number",
			input:  "42",
			target: new(int),
			want:   42,
		},
		{
			name:   "boolean",
			input:  "true",
			target: new(bool),
			want:   true,
		},
		{
			name:   "struct",
			input:  `{"name":"John","age":30,"active":true,"tags":["go","json"],"metadata":{"level":"advanced"}}`,
			target: new(testStruct),
			want: testStruct{
				Name:   "John",
				Age:    30,
				Active: true,
				Tags:   []string{"go", "json"},
				Metadata: map[string]interface{}{
					"level": "advanced",
				},
			},
		},
		{
			name:   "custom unmarshaler",
			input:  `{"custom":"test"}`,
			target: new(customUnmarshaler),
			want:   customUnmarshaler{Value: "test"},
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			target:  new(string),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			
			// Dereference pointer targets for comparison
			switch v := tt.target.(type) {
			case *string:
				assert.Equal(t, tt.want, *v)
			case *int:
				assert.Equal(t, tt.want, *v)
			case *bool:
				assert.Equal(t, tt.want, *v)
			case *testStruct:
				assert.Equal(t, tt.want, *v)
			case *customUnmarshaler:
				assert.Equal(t, tt.want, *v)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	t.Run("simple value", func(t *testing.T) {
		var buf bytes.Buffer
		err := Encode(&buf, "test")
		require.NoError(t, err)
		assert.Equal(t, `"test"`, buf.String())
	})

	t.Run("with json.Marshaler", func(t *testing.T) {
		var buf bytes.Buffer
		err := Encode(&buf, customMarshaler{Value: "encoded"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"custom":"encoded"}`, buf.String())
	})

	t.Run("with io.Reader input", func(t *testing.T) {
		var buf bytes.Buffer
		reader := strings.NewReader(`{"test":"data"}`)
		err := Encode(&buf, reader)
		require.NoError(t, err)
		assert.Equal(t, `{"test":"data"}`, buf.String())
	})

	t.Run("with function encoder", func(t *testing.T) {
		var buf bytes.Buffer
		err := Encode(&buf, func(s *jsoniter.Stream) {
			s.WriteObjectStart()
			s.WriteObjectField("custom")
			s.WriteString("function")
			s.WriteObjectEnd()
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"custom":"function"}`, buf.String())
	})

	t.Run("encode error", func(t *testing.T) {
		var buf bytes.Buffer
		err := Encode(&buf, errorMarshaler{})
		assert.Error(t, err)
	})
}

func TestDecode(t *testing.T) {
	t.Run("simple value", func(t *testing.T) {
		reader := strings.NewReader(`"test"`)
		var result string
		err := Decode(reader, &result)
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("with json.Unmarshaler", func(t *testing.T) {
		reader := strings.NewReader(`{"custom":"decoded"}`)
		var result customUnmarshaler
		err := Decode(reader, &result)
		require.NoError(t, err)
		assert.Equal(t, "decoded", result.Value)
	})

	t.Run("with io.Writer target", func(t *testing.T) {
		reader := strings.NewReader(`{"test":"data"}`)
		var buf bytes.Buffer
		err := Decode(reader, &buf)
		require.NoError(t, err)
		assert.Equal(t, `{"test":"data"}`, buf.String())
	})

	t.Run("decode error", func(t *testing.T) {
		reader := strings.NewReader(`{invalid}`)
		var result string
		err := Decode(reader, &result)
		assert.Error(t, err)
	})
}

func TestMarshalAndEncode(t *testing.T) {
	t.Run("successful marshal and encode", func(t *testing.T) {
		var buf bytes.Buffer
		err := MarshalAndEncode(&buf, map[string]string{"key": "value"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"key":"value"}`, buf.String())
	})

	t.Run("marshal error", func(t *testing.T) {
		var buf bytes.Buffer
		err := MarshalAndEncode(&buf, errorMarshaler{})
		assert.Error(t, err)
	})
}

func TestConfiguration(t *testing.T) {
	t.Run("EscapeHTML is disabled", func(t *testing.T) {
		data, err := Marshal("<script>alert('xss')</script>")
		require.NoError(t, err)
		// Should not escape HTML
		assert.Equal(t, `"<script>alert('xss')</script>"`, string(data))
	})

	t.Run("SortMapKeys is enabled", func(t *testing.T) {
		// Create a map with keys that would have different order
		m := map[string]int{
			"zebra": 1,
			"apple": 2,
			"banana": 3,
		}
		
		data, err := Marshal(m)
		require.NoError(t, err)
		
		// Keys should be sorted alphabetically
		expected := `{"apple":2,"banana":3,"zebra":1}`
		assert.Equal(t, expected, string(data))
	})
}

func TestConcurrency(t *testing.T) {
	// Test that the pool and encoding/decoding work correctly under concurrent access
	t.Run("concurrent marshal", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(n int) {
				defer func() { done <- true }()
				
				data := map[string]int{"value": n}
				result, err := Marshal(data)
				assert.NoError(t, err)
				assert.Contains(t, string(result), strconv.Itoa(n))
			}(i)
		}
		
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Benchmarks
func BenchmarkMarshal(b *testing.B) {
	data := testStruct{
		Name:   "Benchmark",
		Age:    25,
		Active: true,
		Tags:   []string{"test", "benchmark", "json"},
		Metadata: map[string]interface{}{
			"score": 100,
			"level": "expert",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	data := `{"name":"Benchmark","age":25,"active":true,"tags":["test","benchmark","json"],"metadata":{"level":"expert","score":100}}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result testStruct
		err := Unmarshal([]byte(data), &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	data := map[string]string{"key": "value", "test": "benchmark"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := Encode(&buf, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	data := `{"key":"value","test":"benchmark"}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(data)
		var result map[string]string
		err := Decode(reader, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
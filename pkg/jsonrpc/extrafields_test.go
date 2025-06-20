package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtraFields(t *testing.T) {
	t.Run("Set and Get", func(t *testing.T) {
		ef := make(ExtraFields)
		
		ef.Set("key1", "value1")
		ef.Set("key2", 42)
		
		v1, ok1 := ef.Get("key1")
		assert.True(t, ok1)
		assert.Equal(t, "value1", v1)
		
		v2, ok2 := ef.Get("key2")
		assert.True(t, ok2)
		assert.Equal(t, 42, v2)
		
		v3, ok3 := ef.Get("nonexistent")
		assert.False(t, ok3)
		assert.Nil(t, v3)
	})

	t.Run("Del", func(t *testing.T) {
		ef := make(ExtraFields)
		
		ef.Set("key1", "value1")
		v, ok := ef.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", v)
		
		ef.Del("key1")
		v, ok = ef.Get("key1")
		assert.False(t, ok)
		assert.Nil(t, v)
	})

	t.Run("Clone", func(t *testing.T) {
		ef1 := make(ExtraFields)
		ef1.Set("key1", "value1")
		ef1.Set("key2", 42)
		
		ef2 := ef1.Clone()
		
		// Verify cloned values
		v1, ok1 := ef2.Get("key1")
		assert.True(t, ok1)
		assert.Equal(t, "value1", v1)
		
		v2, ok2 := ef2.Get("key2")
		assert.True(t, ok2)
		assert.Equal(t, 42, v2)
		
		// Modify original
		ef1.Set("key1", "modified")
		ef1.Set("key3", "new")
		
		// Clone should not be affected
		v1, ok1 = ef2.Get("key1")
		assert.True(t, ok1)
		assert.Equal(t, "value1", v1)
		
		v3, ok3 := ef2.Get("key3")
		assert.False(t, ok3)
		assert.Nil(t, v3)
		
		// Test nil clone
		var nilEf ExtraFields
		cloned := nilEf.Clone()
		assert.Nil(t, cloned)
	})

	t.Run("Reserved fields cannot be set", func(t *testing.T) {
		ef := make(ExtraFields)
		
		// Try to set reserved fields
		ef.Set("jsonrpc", "3.0")
		ef.Set("id", 999)
		ef.Set("method", "fake")
		ef.Set("params", []string{"test"})
		ef.Set("result", "fake-result")
		ef.Set("error", "fake-error")
		
		// None should be set
		_, ok := ef.Get("jsonrpc")
		assert.False(t, ok)
		_, ok = ef.Get("id")
		assert.False(t, ok)
		_, ok = ef.Get("method")
		assert.False(t, ok)
		_, ok = ef.Get("params")
		assert.False(t, ok)
		_, ok = ef.Get("result")
		assert.False(t, ok)
		_, ok = ef.Get("error")
		assert.False(t, ok)
		assert.Len(t, ef, 0)
		
		// Non-reserved field should work
		ef.Set("custom", "allowed")
		v, ok := ef.Get("custom")
		assert.True(t, ok)
		assert.Equal(t, "allowed", v)
		assert.Len(t, ef, 1)
	})
}
package clientutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

func ptr[T any](v T) *T {
	return &v
}

func TestFillBatch(t *testing.T) {
	msgs := []*jsonrpc.Message{
		{
			ID:     ptr(jsonrpc.ID(`"5"`)),
			Result: jsonrpc.NewStringReader(`["test", "abc", "123"]`),
		},
		{
			ID:     ptr(jsonrpc.ID(`"6"`)),
			Result: jsonrpc.NewStringReader(`12345`),
		},
		{},
		{
			ID:     ptr(jsonrpc.ID(`"7"`)),
			Result: jsonrpc.NewStringReader(`"abcdefgh"`),
		},
	}
	ids := map[int]int{
		0: 5,
		1: 6,
		3: 7,
	}
	b := []*jsonrpc.BatchElem{
		{
			Result: new([]string),
		},
		{
			Result: new(int),
		},
		{},
		{
			Result: new(string),
		},
	}

	FillBatch(ids, msgs, b)

	wantResult := []*jsonrpc.BatchElem{
		{
			Result: &[]string{
				"test",
				"abc",
				"123",
			},
		},
		{
			Result: ptr(12345),
		},
		{},
		{
			Result: ptr("abcdefgh"),
		},
	}

	require.EqualValues(t, len(b), len(wantResult))
	for i := range b {
		expected := wantResult[i]
		actual := b[i]
		assert.EqualValuesf(t, expected.Method, actual.Method, "item %d", i)
		assert.EqualValuesf(t, expected.Result, actual.Result, "item %d", i)
		assert.EqualValuesf(t, expected.Params, actual.Params, "item %d", i)
		assert.EqualValuesf(t, expected.Error, actual.Error, "item %d", i)
	}
}

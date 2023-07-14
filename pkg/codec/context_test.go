package codec

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnContext(t *testing.T) {
	ctx := context.Background()

	{
		_, ok := ConnFromContext(ctx)
		require.False(t, ok)
	}
	d := &DummyClient{}
	ctx = ContextWithConn(ctx, d)

	{
		conn, ok := ConnFromContext(ctx)
		require.True(t, ok)
		require.Equal(t, conn, d)
	}

	{
		conn, ok := StreamingConnFromContext(ctx)
		require.True(t, ok)
		require.Equal(t, conn, d)
	}

}

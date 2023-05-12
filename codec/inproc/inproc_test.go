package inproc_test

import (
	"context"
	"testing"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/codec/inproc"
	"gfx.cafe/open/jrpc/jmux"
	"github.com/stretchr/testify/require"
)

func TestInprocSetup(t *testing.T) {
	mux := jmux.NewMux()
	srv := jrpc.NewServer(mux)

	ctx := context.Background()

	clientCodec := inproc.NewCodec()
	client := inproc.NewClient(clientCodec, nil)
	go func() {
		srv.ServeCodec(ctx, clientCodec)
	}()

	var res any
	err := client.Do(ctx, res, "hi_there", []any{})
	require.ErrorContains(t, err, "does not exist")
}

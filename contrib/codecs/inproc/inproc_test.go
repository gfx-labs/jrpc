package inproc_test

import (
	"context"
	inproc2 "gfx.cafe/open/jrpc/contrib/codecs/inproc"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/pkg/server"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInprocSetup(t *testing.T) {
	mux := jmux.NewMux()
	srv := server.NewServer(mux)

	ctx := context.Background()

	clientCodec := inproc2.NewCodec()
	client := inproc2.NewClient(clientCodec, nil)
	go func() {
		srv.ServeCodec(ctx, clientCodec)
	}()

	var res any
	err := client.Do(ctx, res, "hi_there", []any{})
	require.ErrorContains(t, err, "does not exist")
}

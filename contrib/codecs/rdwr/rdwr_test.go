package rdwr_test

import (
	"context"
	"io"
	"testing"

	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/pkg/server"

	"github.com/stretchr/testify/require"
)

func TestRDWRSetup(t *testing.T) {
	mux := jmux.NewMux()
	srv := server.NewServer(mux)

	ctx := context.Background()

	rd_s, wr_s := io.Pipe()
	rd_c, wr_c := io.Pipe()

	clientCodec := rdwr.NewCodec(rd_s, wr_c, nil)
	client := rdwr.NewClient(rd_c, wr_s, nil)
	go func() {
		srv.ServeCodec(ctx, clientCodec)
	}()

	var res any
	err := client.Do(ctx, res, "hi_there", []any{})
	require.ErrorContains(t, err, "does not exist")
}

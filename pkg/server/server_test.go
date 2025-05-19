package server_test

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"gfx.cafe/open/jrpc/contrib/codecs/rdwr"
	"gfx.cafe/open/jrpc/pkg/jrpctest"
	"gfx.cafe/open/jrpc/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoEthereumTestScripts(t *testing.T) {
	for _, tf := range jrpctest.OriginalTestData.Files {
		t.Run(tf.Name, func(t *testing.T) {
			// create a net pipe
			rd, wr := net.Pipe()
			readbuf := bufio.NewReader(rd)
			srv := jrpctest.NewRouter()
			c := rdwr.NewCodec(wr, wr)
			jsrv := &server.Server{
				BatchParallel: true,
				BatchLimit:    250,
			}
			go jsrv.ServeCodec(context.TODO(), c, srv)
			for _, act := range tf.Action {
				switch act.Direction {
				case jrpctest.DirectionRecv:
					rd.SetReadDeadline(time.Now().Add(5 * time.Second))
					sent, err := readbuf.ReadString('\n')
					require.NoError(t, err)
					assert.EqualValues(t, string(act.Data), strings.TrimSpace(sent))
				case jrpctest.DirectionSend:
					rd.SetWriteDeadline(time.Now().Add(5 * time.Second))
					_, err := rd.Write(append(act.Data, []byte(" ")...))
					require.NoError(t, err)
				}
			}
		})
	}
}

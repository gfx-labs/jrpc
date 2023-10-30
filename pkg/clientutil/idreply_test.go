package clientutil

import (
	"context"
	"io"
	"sync"
	"testing"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"github.com/stretchr/testify/require"
)

const count = 1000

func TestIdReply(t *testing.T) {
	reply := NewIdReply()

	testMessage := "{\"test\": 123}"

	var wg sync.WaitGroup

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			id := reply.NextId()
			v, err := reply.Ask(context.Background(), *id)
			if err != nil {
				t.Error(err)
				return
			}

			x, _ := io.ReadAll(v)
			require.EqualValues(t, testMessage, string(x))
		}()
	}

	for i := 0; i < count; i++ {
		go func(id int) {
			reply.Resolve(jsonrpc.NewNumberID(int64(id+1)), jsonrpc.NewStringReader(testMessage), nil)
		}(i)
	}

	wg.Wait()
}

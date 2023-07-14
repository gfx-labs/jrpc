package clientutil

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"testing"
)

const count = 1000

func TestIdReply(t *testing.T) {
	reply := NewIdReply()

	testMessage := json.RawMessage("{\"test\": 123}")

	var wg sync.WaitGroup

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			id := reply.NextId()
			v, err := reply.Ask(context.Background(), id)
			if err != nil {
				t.Error(err)
				return
			}

			if !bytes.Equal(v, testMessage) {
				t.Error("expected contents to be equal")
				return
			}
		}()
	}

	for i := 0; i < count; i++ {
		go func(id int) {
			reply.Resolve(id+1, testMessage, nil)
		}(i)
	}

	wg.Wait()
}

package clientutil

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/codec"
)

type IdReply struct {
	id atomic.Int64

	chs map[string]chan msgOrError
	mu  sync.Mutex
}

type msgOrError struct {
	msg json.RawMessage
	err error
}

func NewIdReply() *IdReply {
	return &IdReply{
		chs: make(map[string]chan msgOrError, 1),
	}
}

func (i *IdReply) NextId() *codec.ID {
	return codec.NewNumberIDPtr(i.id.Add(1))
}

func (i *IdReply) make(id []byte) <-chan msgOrError {
	i.mu.Lock()
	defer i.mu.Unlock()
	ch := make(chan msgOrError, 1)
	i.chs[string(id)] = ch
	return ch
}

func (i *IdReply) take(id []byte) chan<- msgOrError {
	i.mu.Lock()
	defer i.mu.Unlock()
	ch := i.chs[string(id)]
	delete(i.chs, string(id))
	return ch
}

func (i *IdReply) remove(id []byte) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.chs, string(id))
}

func (i *IdReply) Resolve(id []byte, msg json.RawMessage, err error) {
	ch := i.take(id)
	if ch == nil {
		return
	}

	if err != nil {
		ch <- msgOrError{
			err: err,
		}
		return
	}
	ch <- msgOrError{
		msg: msg,
	}

}

func (i *IdReply) Ask(ctx context.Context, id []byte) (json.RawMessage, error) {
	select {
	case resp := <-i.make(id):
		return resp.msg, resp.err
	case <-ctx.Done():
		i.remove(id)
		return nil, ctx.Err()
	}
}

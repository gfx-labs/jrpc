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

func (i *IdReply) makeOrTake(id []byte) chan msgOrError {
	i.mu.Lock()
	defer i.mu.Unlock()
	if val, ok := i.chs[string(id)]; ok {
		delete(i.chs, string(id))
		return val
	}
	o := make(chan msgOrError)
	i.chs[string(id)] = o
	return o
}

func (i *IdReply) Resolve(id []byte, msg json.RawMessage, err error) {
	if err != nil {
		i.makeOrTake(id) <- msgOrError{
			err: err,
		}
		return
	}
	i.makeOrTake(id) <- msgOrError{
		msg: msg,
	}

}

func (i *IdReply) Ask(ctx context.Context, id []byte) (json.RawMessage, error) {
	select {
	case resp := <-i.makeOrTake(id):
		return resp.msg, resp.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

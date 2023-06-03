package clientutil

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
)

type IdReply struct {
	id atomic.Int64

	chs map[int]chan msgOrError
	mu  sync.Mutex
}

type msgOrError struct {
	msg json.RawMessage
	err error
}

func NewIdReply() *IdReply {
	return &IdReply{
		chs: make(map[int]chan msgOrError, 1),
	}
}

func (i *IdReply) NextId() int {
	return int(i.id.Add(1))
}

func (i *IdReply) makeOrTake(id int) chan msgOrError {
	i.mu.Lock()
	defer i.mu.Unlock()
	if val, ok := i.chs[id]; ok {
		delete(i.chs, id)
		return val
	}
	o := make(chan msgOrError)
	i.chs[id] = o
	return o
}

func (i *IdReply) Resolve(id int, msg json.RawMessage, err error) {
	log.Println(err == nil)
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

func (i *IdReply) Ask(ctx context.Context, id int) (json.RawMessage, error) {
	select {
	case resp := <-i.makeOrTake(id):
		return resp.msg, resp.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

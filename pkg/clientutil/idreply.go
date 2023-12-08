package clientutil

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type IdReply struct {
	id atomic.Int64

	closed chan struct{}

	chs map[string]chan msgOrError
	mu  sync.Mutex
}

type msgOrError struct {
	msg io.ReadCloser
	err error
}

func NewIdReply() *IdReply {
	return &IdReply{
		closed: make(chan struct{}),
		chs:    make(map[string]chan msgOrError, 1),
	}
}

func (i *IdReply) NextId() *jsonrpc.ID {
	return jsonrpc.NewNumberIDPtr(i.id.Add(1))
}

func (i *IdReply) makeOrTake(id []byte) chan msgOrError {
	i.mu.Lock()
	defer i.mu.Unlock()
	ch, ok := i.chs[string(id)]
	if ok {
		// take
		delete(i.chs, string(id))
	} else {
		// make
		ch = make(chan msgOrError, 1)
		i.chs[string(id)] = ch
	}
	return ch

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

func (i *IdReply) Resolve(id []byte, msg io.ReadCloser, err error) {
	ch := i.makeOrTake(id)
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

func (i *IdReply) Ask(ctx context.Context, id []byte) (io.ReadCloser, error) {
	select {
	case resp := <-i.makeOrTake(id):
		return resp.msg, resp.err
	case <-ctx.Done():
		i.remove(id)
		return nil, ctx.Err()
	case <-i.closed:
		return nil, net.ErrClosed
	}
}

func (i *IdReply) Closed() <-chan struct{} {
	return i.closed
}

func (i *IdReply) Close() error {
	close(i.closed)
	return nil
}

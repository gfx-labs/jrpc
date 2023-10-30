package subscription

import (
	"context"
	"strings"
	"sync"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

type Engine struct {
	subscriptions map[SubID]*Notifier
	mu            sync.Mutex
	idgen         func() SubID
}

func NewEngine() *Engine {
	return &Engine{
		subscriptions: make(map[SubID]*Notifier),
		idgen:         randomIDGenerator(),
	}
}

func (e *Engine) closeSub(subid SubID) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	val, ok := e.subscriptions[subid]
	if ok {
		val.err <- ErrSubscriptionClosed
		close(val.err)
		delete(e.subscriptions, subid)
	}
	if !ok {
		return ok, ErrSubscriptionNotFound
	}
	return ok, nil
}

func (e *Engine) Middleware() func(jsonrpc.Handler) jsonrpc.Handler {
	return func(h jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			// its a subscription, so install a notification handler
			switch {
			case strings.HasSuffix(r.Method, serviceMethodSeparator+subscribeMethodSuffix):
				// create the notifier to inject into the context
				n := &Notifier{
					h:         w,
					namespace: strings.TrimSuffix(r.Method, serviceMethodSeparator+subscribeMethodSuffix),
					id:        e.idgen(),
					err:       make(chan error, 1),
				}
				// get the subscription object
				// add to the map
				e.mu.Lock()
				e.subscriptions[n.id] = n
				e.mu.Unlock()
				// now send the subscription id back
				w.Send(n.id, nil)
				// then inject the notifier
				r = r.WithContext(context.WithValue(r.Context(), notifierKey{}, n))
				h.ServeRPC(w, r)
			case strings.HasSuffix(r.Method, serviceMethodSeparator+unsubscribeMethodSuffix):
				// read the subscription id to close
				var subid SubID
				err := r.ParamArray(subid)
				if err != nil {
					w.Send(false, err)
					return
				}
				// close that sub
				w.Send(e.closeSub(subid))
			default:
				h.ServeRPC(w, r)
			}
		})
	}
}

package subscription

import (
	"context"
	"strings"
	"sync"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/pkg/codec"
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

func (e *Engine) Middleware() func(jrpc.Handler) jrpc.Handler {
	return func(h jrpc.Handler) jrpc.Handler {
		return codec.HandlerFunc(func(w codec.ResponseWriter, r *codec.Request) {
			// its a subscription, so install a notification handler
			switch {
			case strings.HasSuffix(r.Method, subscribeMethodSuffix):
				// create the notifier to inject into the context
				n := &Notifier{
					h:         w,
					namespace: strings.TrimSuffix(r.Method, subscribeMethodSuffix),
					id:        e.idgen(),
					err:       make(chan error, 1),
				}
				// get the subscription object
				// add to the map
				e.mu.Lock()
				e.subscriptions[n.id] = n
				e.mu.Unlock()
				// now send the subscription id back
				w.Send(n, nil)
				// then inject the notifier
				r = r.WithContext(context.WithValue(r.Context(), notifierKey{}, n))
				h.ServeRPC(w, r)
			case strings.HasSuffix(r.Method, unsubscribeMethodSuffix):
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

//// handleSubscribe processes *_subscribe method calls.
//func (h *handler) handleSubscribe(cp *callProc, r *Request) *Response {
//	mw := NewReaderResponseWriterMsg(r.WithContext(cp.ctx))
//	switch h.peer.Transport {
//	case "http", "https":
//		mw.Send(nil, ErrNotificationsUnsupported)
//		return mw.Response()
//	}
//
//	// Subscription method name is first argument.
//	name, err := parseSubscriptionName(r.Params)
//	if err != nil {
//		mw.Send(nil, &invalidParamsError{err.Error()})
//		return mw.Response()
//	}
//	namespace := r.namespace()
//	has := h.reg.Match(NewRouteContext(), r.Method)
//	if !has {
//		mw.Send(nil, &subscriptionNotFoundError{namespace, name})
//		return mw.Response()
//	}
//	// Install notifier in context so the subscription handler can find it.
//	n := &Notifier{h: h, namespace: namespace, idgen: randomIDGenerator()}
//	cp.notifiers = append(cp.notifiers, n)
//	// now actually run the handler
//	req := r.WithContext(
//		context.WithValue(r.ctx, notifierKey{}, n),
//	)
//	mw = NewReaderResponseWriterMsg(req)
//	h.reg.ServeRPC(mw, req)
//	return mw.Response()
//}
//
//// parseSubscriptionName extracts the subscription name from an encoded argument array.
//func parseSubscriptionName(rawArgs json.RawMessage) (string, error) {
//	dec := json.NewDecoder(bytes.NewReader(rawArgs))
//	if tok, _ := dec.Token(); tok != json.Delim('[') {
//		return "", errors.New("non-array args")
//	}
//	v, _ := dec.Token()
//	method, ok := v.(string)
//	if !ok {
//		return "", errors.New("expected subscription name as first argument")
//	}
//	return method, nil
//}
//
//func (h *handler) unsubscribe(ctx context.Context, id SubID) (bool, error) {
//	h.subLock.Lock()
//	defer h.subLock.Unlock()
//
//	s := h.serverSubs[id]
//	if s == nil {
//		return false, ErrSubscriptionNotFound
//	}
//	close(s.err)
//	delete(h.serverSubs, id)
//	return true, nil
//}
//
//// cancelServerSubscriptions removes all subscriptions and closes their error channels.
//func (h *handler) cancelServerSubscriptions(err error) {
//	h.subLock.Lock()
//	defer h.subLock.Unlock()
//
//	for id, s := range h.serverSubs {
//		s.err <- err
//		close(s.err)
//		delete(h.serverSubs, id)
//	}
//}
//
//func (h *handler) addSubscriptions(nn []*Notifier) {
//	h.subLock.Lock()
//	defer h.subLock.Unlock()
//
//	for _, n := range nn {
//		if sub := n.takeSubscription(); sub != nil {
//			h.serverSubs[sub.ID] = sub
//		}
//	}
//}

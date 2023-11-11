package subscription

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

var _ jsonrpc.Conn = (*WrapClient)(nil)

type WrapClient struct {
	subs map[string]*clientSub

	conn jsonrpc.Conn
	mu   sync.RWMutex
}

func (w *WrapClient) Closed() <-chan struct{} {
	return w.conn.Closed()
}

func NewWrapClient(conn jsonrpc.Conn) *WrapClient {
	return &WrapClient{
		subs: map[string]*clientSub{},
		conn: conn,
	}
}

func (c *WrapClient) Middleware(h jsonrpc.Handler) jsonrpc.Handler {
	return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		// use normal handler
		if !strings.HasSuffix(r.Method, notificationMethodSuffix) {
			h.ServeRPC(w, r)
			return
		}
		var params subscriptionResult
		// NOTE: this error is ignored because notifications ignore errors
		err := json.Unmarshal(r.Params, &params)
		_ = err
		if params.ID == "" {
			// probably some malformed packet, ignore it
			return
		}
		c.mu.Lock()
		clientSub, ok := c.subs[params.ID]
		c.mu.Unlock()
		if ok {
			clientSub.onmsg <- params.Result
		}
	})
}

func (c *WrapClient) Subscribe(ctx context.Context, namespace string, channel any, args any) (ClientSubscription, error) {
	chanVal := reflect.ValueOf(channel)
	// make sure its a proper channel
	chanVal.Kind()
	if chanVal.Kind() != reflect.Chan || chanVal.Type().ChanDir()&reflect.SendDir == 0 {
		panic("first argument to Subscribe must be a writable channel")
	}
	if chanVal.IsNil() {
		panic("channel given to Subscribe must not be nil")
	}

	// send the actual message to initialize the subscription
	var result string
	err := c.conn.Do(ctx, &result, namespace+serviceMethodSeparator+subscribeMethodSuffix, args)
	if err != nil {
		return nil, err
	}
	// check the result
	if result == "" {
		return nil, ErrSubscriptionNotFound
	}

	// now create a client sub
	sub := &clientSub{
		engine:    c,
		conn:      c.conn,
		namespace: namespace,
		id:        result,
		channel:   chanVal,
		// BUG: a worse is better solution... it means that when this fills, you might receive subscriptions in an undefined error
		onmsg:   make(chan json.RawMessage, 32),
		subdone: make(chan struct{}),
		readErr: make(chan error),
	}

	// will get the type of the event
	etype := chanVal.Type().Elem()

	go func() {
		for {
			select {
			case <-sub.subdone:
				// sub is done, so close readErr
				close(sub.readErr)
			case params, ok := <-sub.onmsg:
				if !ok {
					close(sub.readErr)
					return
				}
				val := reflect.New(etype)
				err := json.Unmarshal(params, val.Interface())
				if err != nil {
					sub.readErr <- err
					return
				}
				// and now send the elem
				sub.channel.Send(val.Elem())
			case <-ctx.Done():
				close(sub.readErr)
				return
			}
		}
	}()

	c.mu.Lock()
	c.subs[sub.id] = sub
	c.mu.Unlock()
	return sub, nil
}
func (c *WrapClient) Do(ctx context.Context, result any, method string, params any) error {
	return c.conn.Do(ctx, result, method, params)
}

func (c *WrapClient) BatchCall(ctx context.Context, b ...*jsonrpc.BatchElem) error {
	return c.conn.BatchCall(ctx, b...)
}

func (c *WrapClient) Close() error {
	return c.conn.Close()
}

func (c *WrapClient) Mount(m jsonrpc.Middleware) {
	c.conn.Mount(m)
}

func (c *WrapClient) Notify(ctx context.Context, method string, params any) error {
	return c.conn.Notify(ctx, method, params)
}

// the actual subscription

type clientSub struct {
	engine    *WrapClient
	conn      jsonrpc.Conn
	namespace string
	id        string
	channel   reflect.Value
	onmsg     chan json.RawMessage
	subdone   chan struct{}

	readErr chan error

	done atomic.Bool
}

func (c *clientSub) Err() <-chan error {
	return c.readErr
}

func (c *clientSub) Unsubscribe() error {
	// TODO: dont use context background here...
	var result string
	err := c.conn.Do(context.Background(), &result, c.namespace+serviceMethodSeparator+unsubscribeMethodSuffix, nil)
	if err != nil {
		return err
	}
	if c.done.CompareAndSwap(false, true) {
		close(c.subdone)
	}
	return nil
}

func (c *clientSub) String() string {
	return c.id
}

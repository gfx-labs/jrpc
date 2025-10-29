package subscription

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
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
		err := sonic.ConfigStd.Unmarshal(r.Params, &params)
		_ = err
		if params.ID == "" {
			// probably some malformed packet, ignore it
			return
		}
		c.mu.RLock()
		clientSub, ok := c.subs[params.ID]
		c.mu.RUnlock()
		if ok {
			clientSub.notify(params.Result)
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

	// FIXME(garet): we can lose subscription messages here. if they send a notification and the server handles it
	// before we finish adding it to the subs, the message will be lost. Ping will almost always be much much longer
	// than us adding the sub so it probably doesn't matter. but it fails the unit tests :(

	// now create a client sub
	sub := &clientSub{
		engine:    c,
		conn:      c.conn,
		namespace: namespace,
		id:        result,
		channel:   chanVal,
		readErr:   make(chan error, 1),
	}

	go func() {
		defer func() {
			_ = sub.Unsubscribe()
		}()
		select {
		case <-c.Closed():
		case <-ctx.Done():
			sub.err(ctx.Err())
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

	channel reflect.Value

	readErr chan error
	closed  bool
	mu      sync.Mutex
}

func (c *clientSub) err(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	select {
	case c.readErr <- err:
	default:
	}
}

func (c *clientSub) notify(result json.RawMessage) {
	val := reflect.New(c.channel.Type().Elem())
	err := sonic.ConfigStd.Unmarshal(result, val.Interface())
	if err != nil {
		c.err(err)
		return
	}
	reflect.Select([]reflect.SelectCase{
		{
			Dir:  reflect.SelectSend,
			Chan: c.channel,
			Send: val.Elem(),
		},
		{
			Dir: reflect.SelectDefault,
		},
	})
}

func (c *clientSub) Err() <-chan error {
	return c.readErr
}

func (c *clientSub) Unsubscribe() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.readErr)
	c.mu.Unlock()

	c.engine.mu.Lock()
	delete(c.engine.subs, c.id)
	c.engine.mu.Unlock()

	// TODO: dont use context background here...
	var result bool
	err := c.conn.Do(context.Background(), &result, c.namespace+serviceMethodSeparator+unsubscribeMethodSuffix, []string{
		c.id,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientSub) String() string {
	return c.id
}

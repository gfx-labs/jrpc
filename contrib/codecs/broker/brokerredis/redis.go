package redis

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"gfx.cafe/open/jrpc/contrib/codecs/broker"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
)

type Broker struct {
	client redis.UniversalClient
	domain string

	id xid.ID
}

type subscription struct {
	ch  chan json.RawMessage
	err error

	closed atomic.Bool
	mu     sync.RWMutex
	pubsub *redis.PubSub
}

// channel that will close when done or error
func (s *subscription) Listen() <-chan json.RawMessage {
	return s.ch
}

// should close the channel and also stop listening
func (s *subscription) Close() error {
	if s.closed.CompareAndSwap(false, true) {
		s.pubsub.Close()
	}
	return nil
}

// this hold errors
func (s *subscription) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}

func (s *Broker) WriteRequest(ctx context.Context, clientId string, msg json.RawMessage) error {
	req, err := json.Marshal(&RedisRequest{
		ReplyChannel: clientId,
		Message:      msg,
	})
	if err != nil {
		return err
	}
	return s.client.LPush(ctx, s.domain+reqDomainSuffix, req).Err()
}

func (s *Broker) Subscribe(ctx context.Context, clientId string) (broker.Subscription, error) {
	topic := s.domain + "." + clientId + respDomainSuffix
	sub := &subscription{
		ch: make(chan json.RawMessage, 16),
	}

	sub.pubsub = s.client.Subscribe(ctx, topic)
	ch := sub.pubsub.Channel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				sub.Close()
				return
			case t := <-ch:
				select {
				case sub.ch <- json.RawMessage(stringToBytes(t.Payload)):
				default:
				}
			}
		}
	}()

	return sub, nil
}

func CreateBroker(ctx context.Context, domain string, opts *redis.UniversalOptions) *Broker {
	c := redis.NewUniversalClient(opts)
	// the xid doesn't need to be secure, since we assume anyone with access to the redis cluster can do anything anyways.
	s := &Broker{
		client: c,
		id:     xid.New(),
		domain: domain,
	}
	return s
}

type RedisRequest struct {
	ReplyChannel string          `json:"r"`
	Message      json.RawMessage `json:"msg"`
}

func (s *Broker) ReadRequest(ctx context.Context) (json.RawMessage, func(json.RawMessage) error, error) {
	timeout := time.Second
	res, err := s.client.BLPop(ctx, timeout, s.domain+reqDomainSuffix).Result()
	if err != nil {
		return nil, nil, err
	}
	if len(res) != 2 {
		return nil, nil, err
	}
	redisReq := &RedisRequest{}
	err = json.Unmarshal(stringToBytes(res[1]), redisReq)
	if err != nil {
		return nil, nil, err
	}
	return redisReq.Message, func(rm json.RawMessage) error {
		if len(rm) == 0 {
			return nil
		}
		target := s.domain + "." + redisReq.ReplyChannel + respDomainSuffix
		err := s.client.Publish(context.Background(), target, []byte(rm)).Err()
		if err != nil {
			return err
		}
		return nil
	}, nil
}

const reqDomainSuffix = ".req"
const respDomainSuffix = ".resp"

// stringHeader is the runtime representation of a string.
// It should be identical to reflect.StringHeader
type stringHeader struct {
	data      unsafe.Pointer
	stringLen int
}

// sliceHeader is the runtime representation of a slice.
// It should be identical to reflect.sliceHeader
type sliceHeader struct {
	data     unsafe.Pointer
	sliceLen int
	sliceCap int
}

func stringToBytes(s string) (b []byte) {
	stringHeader := (*stringHeader)(unsafe.Pointer(&s))
	sliceHeader := (*sliceHeader)(unsafe.Pointer(&b))
	sliceHeader.data = stringHeader.data
	sliceHeader.sliceLen = len(s)
	sliceHeader.sliceCap = len(s)
	return b
}

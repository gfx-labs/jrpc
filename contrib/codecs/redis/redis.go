package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"unsafe"

	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
)

type ServerStream struct {
	client redis.UniversalClient
	domain string

	id xid.ID
}

func CreateServerStream(ctx context.Context, domain string, opts *redis.UniversalOptions) (*ServerStream, error) {
	c := redis.NewUniversalClient(opts)
	// the xid doesn't need to be secure, since we assume anyone with access to the redis cluster can do anything anyways.
	s := &ServerStream{
		client: c,
		id:     xid.New(),
		domain: domain,
	}
	return s, nil
}

type RedisRequest struct {
	ReplyChannel string          `json:"r"`
	Message      json.RawMessage `json:"msg"`
}

func (s *ServerStream) ReadRequest(ctx context.Context) (*RedisRequest, func(json.RawMessage) error, error) {
	timeout := time.Hour
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
	log.Println("got req", redisReq.ReplyChannel, string(redisReq.Message))
	return redisReq, func(rm json.RawMessage) error {
		target := s.domain + "." + redisReq.ReplyChannel
		log.Println("replying", target, string(rm))
		return s.client.Publish(context.Background(), target, string(rm)).Err()
	}, nil
}

const reqDomainSuffix = ".req"
const respDomainSuffix = ".resp"

func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

package broker

import (
	"context"
	"encoding/json"
	"io"
)

type ServerSpoke interface {
	ReadRequest(ctx context.Context) (json.RawMessage, Replier, error)
}

type ClientSpoke interface {
	WriteRequest(ctx context.Context, clientId string, msg json.RawMessage) error
	Subscribe(ctx context.Context, clientId string) (Subscription, error)
}

type Broker interface {
	ServerSpoke
	ClientSpoke
}

type Replier interface {
	Send(fn func(io.Writer) error) error
}

type ReplierFunc func(fn func(io.Writer) error) error

func (r ReplierFunc) Send(fn func(io.Writer) error) error {
	return r(fn)
}

type Subscription interface {
	// channel that will close when done or error
	Listen() <-chan json.RawMessage
	// should close the channel and also stop listening
	Close() error
	// this hold errors
	Err() error
}

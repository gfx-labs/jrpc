package broker

import (
	"context"
	"encoding/json"
)

type ServerSpoke interface {
	ReadRequest(ctx context.Context) (json.RawMessage, func(json.RawMessage) error, error)
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
	Send(json.RawMessage)
}

type Subscription interface {
	// channel that will close when done or error
	Listen() <-chan json.RawMessage
	// should close the channel and also stop listening
	Close() error
	// this hold errors
	Err() error
}

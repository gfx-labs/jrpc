package broker

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
)

type subscription struct {
	topic string
	ch    chan json.RawMessage
	err   error

	closed atomic.Bool
	mu     sync.RWMutex
}

// channel that will close when done or error
func (s *subscription) Listen() <-chan json.RawMessage {
	return s.ch
}

// should close the channel and also stop listening
func (s *subscription) Close() error {
	s.closed.CompareAndSwap(false, true)
	return nil
}

// this hold errors
func (s *subscription) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}

type frame struct {
	topic string
	data  json.RawMessage
}

type ChannelBroker struct {
	mu       sync.RWMutex
	subs     map[int]*subscription
	subCount int

	onDroppedMessage func(string, []byte)

	msgs chan *frame

	domain string
}

func NewChannelBroker() *ChannelBroker {
	return &ChannelBroker{
		subs: map[int]*subscription{},
		msgs: make(chan *frame, 128),
	}
}

func (b *ChannelBroker) SetDroppedMessageHandler(fn func(string, []byte)) *ChannelBroker {
	b.onDroppedMessage = fn
	return b
}

func (b *ChannelBroker) ReadRequest(ctx context.Context) (json.RawMessage, func(json.RawMessage) error, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case f := <-b.msgs:
		return f.data, func(resp json.RawMessage) error {
			return b.Publish(context.Background(), f.topic, resp)
		}, nil
	}
}

func (b *ChannelBroker) WriteRequest(ctx context.Context, topic string, msg json.RawMessage) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case b.msgs <- &frame{data: msg, topic: topic}:
	}
	return nil
}

func (b *ChannelBroker) Publish(ctx context.Context, topic string, data []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, v := range b.subs {
		if v.closed.Load() {
			continue
		}
		if childTopicMatchesParent(v.topic, topic) {
			select {
			case v.ch <- data:
			default:
				if b.onDroppedMessage != nil {
					b.onDroppedMessage(topic, data)
				}
			}
		}
	}
	return nil
}

func (b *ChannelBroker) Subscribe(ctx context.Context, topic string) (Subscription, error) {
	sub := &subscription{
		topic: topic,
		ch:    make(chan json.RawMessage, 16),
	}
	b.mu.Lock()
	b.subCount = b.subCount + 1
	id := b.subCount
	b.subs[id] = sub
	b.mu.Unlock()
	// gc after adding a new subscription
	b.gc()
	return sub, nil
}

func (b *ChannelBroker) gc() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for k, v := range b.subs {
		if v.closed.Load() {
			delete(b.subs, k)
		}
	}
}

// This is to see if a message with topic child should be matched with subscription parent
// so the child cannot contain wildcards * and >, but the parent can
func childTopicMatchesParent(parentString string, childString string) bool {
	parent := strings.Split(parentString, ".")
	child := strings.Split(childString, ".")
	// if the length of the child is less than the length of the parent, its not possible for match
	// for instance, if the parent topic is "one.two", and the child is "one", it will never match
	if len(child) < len(parent) {
		return false
	}
	// this is safe because length of child must be equal to or lower than parent, from previous
	for idx, v := range parent {
		// if parent is wildcard, match all, so continue
		if v == "*" {
			continue
		}
		// if the > wildcard is at the end, and we have exited since then, we are done
		if v == ">" && len(parent)-1 == idx {
			return true
		}
		// else make sure parent matches child.
		if child[idx] != v {
			return false
		}
	}
	if len(child) == len(parent) {
		return true
	}
	return false
}

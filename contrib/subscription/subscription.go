package subscription

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/util/go/frand"

	json "github.com/goccy/go-json"
)

const (
	subscribeMethodSuffix    = "_subscribe"
	notificationMethodSuffix = "_subscription"
	unsubscribeMethodSuffix  = "_unsubscribe"
	serviceMethodSeparator   = "_"

	maxClientSubscriptionBuffer = 12800
)

var (
	// ErrNotificationsUnsupported is returned when the connection doesn't support notifications
	ErrNotificationsUnsupported = errors.New("notifications not supported")
	// ErrNotificationNotFound is returned when the notification for the given id is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

var globalInc = atomic.Int64{}
var globalGen = randomIDGenerator()

type SubID string

// NewID returns a new, random ID.
func NewID() SubID {
	return globalGen()
}

// randomIDGenerator returns a function generates a random IDs.
func randomIDGenerator() func() SubID {
	return func() SubID {
		id := make([]byte, 32)
		frand.Read(id)
		id = binary.LittleEndian.AppendUint64(id, uint64(globalInc.Add(1)))
		return encodeSubID(id)
	}
}

func encodeSubID(b []byte) SubID {
	id := hex.EncodeToString(b)
	id = strings.TrimLeft(id, "0")
	if id == "" {
		id = "0" // ID's are RPC quantities, no leading zero's and 0 is 0x0.
	}
	return SubID("0x" + id)
}

type subscriptionResult struct {
	ID     string          `json:"subscription"`
	Result json.RawMessage `json:"result,omitempty"`
}

type notifierKey struct{}

// NotifierFromContext returns the Notifier value stored in ctx, if any.
func NotifierFromContext(ctx context.Context) (*Notifier, bool) {
	n, ok := ctx.Value(notifierKey{}).(*Notifier)
	return n, ok
}

// Notifier is tied to a RPC connection that supports subscriptions.
// Server callbacks use the notifier to send notifications.
type Notifier struct {
	h         codec.ResponseWriter
	namespace string

	mu  sync.Mutex
	sub *Subscription

	id SubID
}

// CreateSubscription returns a new subscription that is coupled to the
// RPC connection. By default subscriptions are inactive and notifications
// are dropped until the subscription is marked as active. This is done
// by the RPC server after the subscription ID is send to the client.
func (n *Notifier) createSubscription() *Subscription {
	n.sub = &Subscription{
		ID:        n.id,
		namespace: n.namespace,
		err:       make(chan error, 1),
	}
	return n.sub
}

// Notify sends a notification to the client with the given data as payload.
// If an error occurs the RPC connection is closed and the error is returned.
func (n *Notifier) Notify(data interface{}) error {
	enc, err := json.Marshal(data)
	if err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.send(n.sub, enc)
}

func (n *Notifier) send(sub *Subscription, data json.RawMessage) error {
	params, _ := json.Marshal(&subscriptionResult{ID: string(sub.ID), Result: data})
	return n.h.Notify(n.namespace+notificationMethodSuffix, json.RawMessage(params))
}

// A Subscription is created by a notifier and tied to that notifier. The client can use
// this subscription to wait for an unsubscribe request for the client, see Err().
type Subscription struct {
	ID        SubID
	namespace string
	err       chan error // closed on unsubscribe
}

// Err returns a channel that is closed when the client send an unsubscribe request.
func (s *Subscription) Err() <-chan error {
	return s.err
}

// MarshalJSON marshals a subscription as its ID.
func (s *Subscription) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ID)
}

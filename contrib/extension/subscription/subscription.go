package subscription

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"gfx.cafe/open/jrpc/pkg/jjson"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

var serviceMethodSeparator = "/"

func SetServiceMethodSeparator(val string) {
	serviceMethodSeparator = val
}

const (
	subscribeMethodSuffix    = "subscribe"
	notificationMethodSuffix = "subscription"
	unsubscribeMethodSuffix  = "unsubscribe"

	maxClientSubscriptionBuffer = 12800
)

var (
	// ErrNotificationsUnsupported is returned when the connection doesn't support notifications
	ErrNotificationsUnsupported = errors.New("notifications not supported")
	// ErrNotificationNotFound is returned when the notification for the given id is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
	// ErrNotificationNotFound is returned when the notification for the given id is not found
	ErrSubscriptionClosed = errors.New("subscription closed. not found")
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
		rand.Read(id)
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
	h         jsonrpc.ResponseWriter
	namespace string

	mu sync.Mutex

	id  SubID
	err chan error

	sentId bool
}

func (n *Notifier) ID() SubID {
	return n.id
}

// Notify sends a notification to the client with the given data as payload.
// If an error occurs the RPC connection is closed and the error is returned.
func (n *Notifier) Notify(data any) error {
	enc, err := jjson.Marshal(data)
	if err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.send(enc)
}

func (n *Notifier) Err() <-chan error {
	return n.err
}

func (n *Notifier) send(data json.RawMessage) error {
	params, _ := jjson.Marshal(&subscriptionResult{ID: string(n.id), Result: data})
	// try to send the id back. this will just fail with errAlreadySent if its already been sent.
	// so it is safe-ish to just ignore this error
	// technically we should check for jsonrpc.ErrSendAlreadyCalled and then error earlier otherwise... but is that really right?
	if n.sentId == false {
		_ = n.h.Send(n.id, nil)
		n.sentId = true
	}

	err := n.h.Notify(
		n.namespace+
			serviceMethodSeparator+
			notificationMethodSuffix, json.RawMessage(params))

	if err != nil {
		n.err <- err
		return err
	}
	return nil
}

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
	// ErrNotificationNotFound is returned when the notification for the given id is not found
	ErrSubscriptionClosed = errors.New("subscription not found")
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

	mu sync.Mutex

	id  SubID
	err chan error // closed on unsubscribe
}

// Notify sends a notification to the client with the given data as payload.
// If an error occurs the RPC connection is closed and the error is returned.
func (n *Notifier) Notify(data any) error {
	enc, err := json.Marshal(data)
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
	params, _ := json.Marshal(&subscriptionResult{ID: string(n.id), Result: data})
	return n.h.Notify(n.namespace+notificationMethodSuffix, json.RawMessage(params))
}

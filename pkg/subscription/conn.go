package subscription

import (
	"context"

	"gfx.cafe/open/jrpc"
)

type SubscriptionConn interface {
	jrpc.StreamingConn

	Subscribe(ctx context.Context, namespace string, channel any, args ...any) (*ClientSubscription, error)
}

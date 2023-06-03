package subscription

import (
	"context"
	"gfx.cafe/open/jrpc/pkg/codec"
)

type SubscriptionConn interface {
	codec.StreamingConn

	Subscribe(ctx context.Context, namespace string, channel any, args ...any) (*ClientSubscription, error)
}

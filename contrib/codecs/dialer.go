package codecs

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

var ErrSchemeNotSupported = errors.New("url scheme not supported")

func DialContext(ctx context.Context, u string) (jsonrpc.Conn, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	dialer := dialers[pu.Scheme]
	if dialer == nil {
		return nil, fmt.Errorf("%w: %s", ErrSchemeNotSupported, pu.Scheme)
	}
	return dialer(ctx, u)
}

func Dial(u string) (jsonrpc.Conn, error) {
	ctx := context.Background()
	return DialContext(ctx, u)
}

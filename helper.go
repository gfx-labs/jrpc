package jrpc

import "context"

func Do[T any](ctx context.Context, c Conn, method string, args any) (T, error) {
	var t T
	err := c.Do(ctx, t, method, args)
	if err != nil {
		return t, err
	}
	return t, nil
}

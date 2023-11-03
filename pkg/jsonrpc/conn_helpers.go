package jsonrpc

import "context"

// Do
func Do[T any](ctx context.Context, c Conn, method string, args any) (*T, error) {
	var t T
	err := c.Do(ctx, &t, method, args)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Call
func Call[T any](ctx context.Context, c Conn, method string, args ...any) (*T, error) {
	var t T
	err := c.Do(ctx, &t, method, args)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CallInto
func CallInto(ctx context.Context, c Conn, result any, method string, args ...any) error {
	err := c.Do(ctx, result, method, args)
	if err != nil {
		return err
	}
	return nil
}

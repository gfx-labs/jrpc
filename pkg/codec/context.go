package codec

import "context"

type clientContextKey struct{}

// ClientFromContext retrieves the client from the context, if any. This can be used to perform
// 'reverse calls' in a handler method.
func ContextWithConn(ctx context.Context, c Conn) context.Context {
	client, _ := ctx.Value(clientContextKey{}).(Conn)
	return context.WithValue(ctx, clientContextKey{}, client)
}

// ClientFromContext retrieves the client from the context, if any. This can be used to perform
// 'reverse calls' in a handler method.
func ConnFromContext(ctx context.Context) (Conn, bool) {
	client, ok := ctx.Value(clientContextKey{}).(Conn)
	return client, ok
}

func StreamingConnFromContext(ctx context.Context) (StreamingConn, bool) {
	client, ok := ctx.Value(clientContextKey{}).(StreamingConn)
	return client, ok
}

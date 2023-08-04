package codec

import "context"

var _ Conn = (*DummyClient)(nil)

type DummyClient struct{}

func (d *DummyClient) Closed() <-chan struct{} {
	panic("not implemented") // TODO: Implement
}

func (d *DummyClient) Notify(ctx context.Context, method string, params any) error {
	panic("not implemented") // TODO: Implement
}

func (d *DummyClient) Do(ctx context.Context, result any, method string, params any) error {
	panic("not implemented") // TODO: Implement
}

func (d *DummyClient) BatchCall(ctx context.Context, b ...*BatchElem) error {
	panic("not implemented") // TODO: Implement
}

func (d *DummyClient) Close() error {
	panic("not implemented") // TODO: Implement
}

func (d *DummyClient) Mount(_ Middleware) {
	panic("not implemented") // TODO: Implement
}

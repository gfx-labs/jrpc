package jrpc

func Do[T any](c *Client, method string, args any) (T, error) {
	var t T
	err := c.Do(t, method, args)
	if err != nil {
		return t, err
	}
	return t, nil
}

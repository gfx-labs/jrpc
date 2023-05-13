// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package jrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/codec"
)

var (
	ErrClientQuit                = errors.New("client is closed")
	ErrNoResult                  = errors.New("no result in JSON-RPC response")
	ErrSubscriptionQueueOverflow = errors.New("subscription queue overflow")
	errClientReconnected         = errors.New("client reconnected")
	errDead                      = errors.New("connection lost")
)

const (
	// Timeouts
	defaultDialTimeout = 10 * time.Second // used if context has no deadline
	subscribeTimeout   = 5 * time.Second  // overall timeout eth_subscribe, rpc_modules calls
)

// Client represents a connection to an RPC server.
type Client struct {
	remote string
	c      *http.Client

	id atomic.Int64
}

func Dial(ctx context.Context, client *http.Client, target string) (*Client, error) {
	return &Client{remote: target, c: client}, nil
}

func (c *Client) Do(ctx context.Context, result any, method string, params any) error {
	req := jrpc.NewRequestInt(ctx, int(c.id.Add(1)), method, params)
	dat, err := req.MarshalJSON()
	if err != nil {
		return err
	}
	resp, err := c.c.Post(c.remote, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		json.NewDecoder(resp.Body).Decode(&result)
	}
	return nil
}

func (c *Client) Notify(ctx context.Context, result any, method string, params any) error {
	req := jrpc.NewRequestInt(ctx, int(c.id.Add(1)), method, params)
	dat, err := req.MarshalJSON()
	if err != nil {
		return err
	}
	_, err = c.c.Post(c.remote, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return err
	}
	return err
}

func (c *Client) BatchCall(ctx context.Context, b ...*jrpc.BatchElem) error {
	reqs := make([]*jrpc.Request, len(b))
	ids := make([]int, 0, len(b))
	for _, v := range b {
		if v.IsNotification {
			reqs = append(reqs, jrpc.NewRequest(ctx, "", v.Method, v.Params))
		} else {
			id := int(c.id.Add(1))
			ids = append(ids, id)
			reqs = append(reqs, jrpc.NewRequestInt(ctx, id, v.Method, v.Params))
		}
	}
	dat, err := json.Marshal(b)
	if err != nil {
		return err
	}
	resp, err := c.c.Post(c.remote, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	msgs := []*codec.Message{}
	err = json.NewDecoder(resp.Body).Decode(&msgs)
	if err != nil {
		return err
	}
	answers := map[int]*codec.Message{}
	for _, v := range msgs {
		answers[v.ID.Number()] = v
	}

	for i := range ids {
		idx := i
		ans, ok := answers[i]
		if !ok {
			b[idx].Error = fmt.Errorf("No response found")
			continue
		}
		if b[idx].Result == nil {
			continue
		}
		err = json.Unmarshal(ans.Result, b[idx].Result)
		if err != nil {
			b[idx].Error = err
		}
	}
	return nil
}

func (c *Client) Close() error {
	return nil
}

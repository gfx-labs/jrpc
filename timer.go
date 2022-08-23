package jrpc

import (
	"time"
)

type Timer struct {
	s time.Time
}

func NewTimer() *Timer {
	return &Timer{
		s: time.Now(),
	}
}

func (t *Timer) Since(...any) time.Duration {
	return time.Now().Sub(t.s)
}

func (t *Timer) Until() time.Duration {
	return t.s.Sub(time.Now())
}

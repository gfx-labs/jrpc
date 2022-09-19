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
	return time.Since(t.s)
}

func (t *Timer) Until() time.Duration {
	return time.Until(t.s)
}

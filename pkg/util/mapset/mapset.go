package mapset

import "sync"

type Set[T comparable] struct {
	m  map[T]struct{}
	mu sync.RWMutex
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		m: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(x T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[x] = struct{}{}
}

func (s *Set[T]) Remove(x T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, x)
}

func (s *Set[T]) Each(fn func(x T) bool) {
	for k := range s.m {
		if !fn(k) {
			return
		}
	}
}

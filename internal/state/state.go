package state

import (
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]any),
	}
}

func (s *Store) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *Store) All() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c := make(map[string]any, len(s.data))
	for k, v := range s.data {
		c[k] = v
	}
	return c
}

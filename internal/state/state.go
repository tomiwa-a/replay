package state

import (
	"sync"
	"time"
)

type ScopeType int

const (
	ScopeGlobal   ScopeType = iota
	ScopeWorkflow
	ScopeStep
)

type entry struct {
	value any
	ttl   time.Duration
	born  time.Time
}

type scope struct {
	typ   ScopeType
	data  map[string]any
	inner map[string]time.Duration
}

type Store struct {
	mu     sync.RWMutex
	scopes []*scope
}

func NewStore() *Store {
	s := &Store{
		scopes: make([]*scope, 0, 8),
	}
	s.scopes = append(s.scopes, &scope{
		typ:   ScopeGlobal,
		data:  make(map[string]any),
		inner: make(map[string]time.Duration),
	})
	return s
}

func (s *Store) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scopes[len(s.scopes)-1].data[key] = val
}

func (s *Store) SetTTL(key string, val any, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	top := s.scopes[len(s.scopes)-1]
	top.data[key] = entry{value: val, ttl: ttl, born: time.Now()}
}

func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if v, ok := s.scopes[i].data[key]; ok {
			if e, isEntry := v.(entry); isEntry {
				if e.ttl > 0 && time.Since(e.born) > e.ttl {
					delete(s.scopes[i].data, key)
					continue
				}
				return e.value, true
			}
			return v, true
		}
	}
	return nil, false
}

func (s *Store) MustGet(key string) any {
	v, _ := s.Get(key)
	return v
}

func (s *Store) All() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	c := make(map[string]any, 64)
	for i := len(s.scopes) - 1; i >= 0; i-- {
		for k, v := range s.scopes[i].data {
			if e, isEntry := v.(entry); isEntry {
				if e.ttl > 0 && now.Sub(e.born) > e.ttl {
					continue
				}
				c[k] = e.value
			} else {
				c[k] = v
			}
		}
	}
	return c
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if _, ok := s.scopes[i].data[key]; ok {
			delete(s.scopes[i].data, key)
			return
		}
	}
}

func (s *Store) Enter(typ ScopeType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scopes = append(s.scopes, &scope{
		typ:   typ,
		data:  make(map[string]any),
		inner: make(map[string]time.Duration),
	})
}

func (s *Store) Exit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.scopes) > 1 {
		s.scopes = s.scopes[:len(s.scopes)-1]
	}
}

func (s *Store) ExitUntil(typ ScopeType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for len(s.scopes) > 1 && s.scopes[len(s.scopes)-1].typ != typ {
		s.scopes = s.scopes[:len(s.scopes)-1]
	}
}

func (s *Store) Snapshot() map[string]any {
	return s.All()
}

func (s *Store) Restore(snapshot map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	top := s.scopes[len(s.scopes)-1]
	top.data = make(map[string]any, len(snapshot))
	for k, v := range snapshot {
		top.data[k] = v
	}
}

// Promote copies all values from the current (step) scope into the
// parent (workflow) scope. This makes extracted values visible to
// subsequent steps after the step scope exits.
func (s *Store) Promote() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.scopes) < 2 {
		return
	}
	top := s.scopes[len(s.scopes)-1]
	parent := s.scopes[len(s.scopes)-2]
	for k, v := range top.data {
		parent.data[k] = v
	}
}

func (s *Store) ScopeDepth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.scopes)
}

func (s *Store) CurrentScope() ScopeType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scopes[len(s.scopes)-1].typ
}

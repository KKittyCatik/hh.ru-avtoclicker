package httptransport

import (
	"sync"
	"time"
)

type stateStore struct {
	mu     sync.Mutex
	states map[string]time.Time
}

func newStateStore() *stateStore {
	return &stateStore{states: make(map[string]time.Time)}
}

func (s *stateStore) Set(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(10 * time.Minute)
}

func (s *stateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.states[state]
	if !ok {
		return false
	}
	delete(s.states, state)
	return time.Now().Before(exp)
}

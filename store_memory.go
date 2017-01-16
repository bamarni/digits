package digits

import (
	"sync"
)

type MemoryStore struct {
	idts map[string]*Identity
	mu   sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{idts: make(map[string]*Identity)}
}

func (s *MemoryStore) Load(creds string) *Identity {
	s.mu.RLock()
	idt := s.idts[creds]
	s.mu.RUnlock()
	return idt
}

func (s *MemoryStore) Save(creds string, idt *Identity) {
	s.mu.Lock()
	s.idts[creds] = idt
	s.mu.Unlock()
}

// Package store owns the application's in-memory state.
package store

import (
	"sync"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
)

// MemoryStore keeps accounts in memory and is safe for concurrent access.
type MemoryStore struct {
	mu       sync.RWMutex
	accounts map[string]account.Account
}

// NewMemoryStore creates an empty account store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		accounts: make(map[string]account.Account),
	}
}

// Find returns an account and whether it exists.
func (s *MemoryStore) Find(id string) (account.Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storedAccount, found := s.accounts[id]
	return storedAccount, found
}

// Save creates or replaces an account.
func (s *MemoryStore) Save(accountToSave account.Account) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts[accountToSave.ID] = accountToSave
}

// Reset removes all accounts from the store.
func (s *MemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts = make(map[string]account.Account)
}

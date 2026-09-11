package account

import "sync"

// Store defines the state operations required by the account service.
type Store interface {
	Find(id string) (Account, bool)
	Save(account Account)
}

// Service applies account business rules.
type Service struct {
	mu    sync.Mutex
	store Store
}

// NewService creates an account service backed by the provided store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Deposit creates or credits the destination account.
func (s *Service) Deposit(destination string, amount int64) (Account, error) {
	if amount <= 0 {
		return Account{}, ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	destinationAccount, found := s.store.Find(destination)
	if !found {
		destinationAccount = Account{ID: destination}
	}

	destinationAccount.Balance += amount
	s.store.Save(destinationAccount)

	return destinationAccount, nil
}

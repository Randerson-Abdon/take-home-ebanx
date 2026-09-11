package account

import "sync"

// Store defines the state operations required by the account service.
type Store interface {
	Find(id string) (Account, bool)
	Save(account Account)
	SaveAll(accounts ...Account)
	Reset()
}

// Service applies account business rules.
type Service struct {
	mu    sync.RWMutex
	store Store
}

// NewService creates an account service backed by the provided store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Balance returns the current balance without changing account state.
func (s *Service) Balance(id string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storedAccount, found := s.store.Find(id)
	if !found {
		return 0, ErrAccountNotFound
	}

	return storedAccount.Balance, nil
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

// Withdraw debits an existing account when it has sufficient funds.
func (s *Service) Withdraw(origin string, amount int64) (Account, error) {
	if amount <= 0 {
		return Account{}, ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	originAccount, found := s.store.Find(origin)
	if !found {
		return Account{}, ErrAccountNotFound
	}
	if originAccount.Balance < amount {
		return Account{}, ErrInsufficientFunds
	}

	originAccount.Balance -= amount
	s.store.Save(originAccount)

	return originAccount, nil
}

// Transfer moves funds from an existing origin to a destination account.
func (s *Service) Transfer(origin, destination string, amount int64) (Account, Account, error) {
	if amount <= 0 {
		return Account{}, Account{}, ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	originAccount, found := s.store.Find(origin)
	if !found {
		return Account{}, Account{}, ErrAccountNotFound
	}
	if originAccount.Balance < amount {
		return Account{}, Account{}, ErrInsufficientFunds
	}
	if origin == destination {
		return originAccount, originAccount, nil
	}

	destinationAccount, found := s.store.Find(destination)
	if !found {
		destinationAccount = Account{ID: destination}
	}

	originAccount.Balance -= amount
	destinationAccount.Balance += amount
	s.store.SaveAll(originAccount, destinationAccount)

	return originAccount, destinationAccount, nil
}

// Reset removes all account state.
func (s *Service) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store.Reset()
}

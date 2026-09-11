package account_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/store"
)

func TestBalanceReturnsExistingAccountBalance(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	memoryStore.Save(account.Account{ID: "100", Balance: 20})
	service := account.NewService(memoryStore)

	got, err := service.Balance("100")

	if err != nil {
		t.Fatalf("balance returned an unexpected error: %v", err)
	}
	if got != 20 {
		t.Fatalf("expected balance 20, got %d", got)
	}
}

func TestBalanceReturnsErrorForMissingAccount(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	service := account.NewService(memoryStore)

	got, err := service.Balance("missing")

	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
	if got != 0 {
		t.Fatalf("expected zero balance, got %d", got)
	}
	if _, found := memoryStore.Find("missing"); found {
		t.Fatal("expected balance lookup not to create an account")
	}
}

func TestBalanceDoesNotChangeAccountState(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	want := account.Account{ID: "100", Balance: 20}
	memoryStore.Save(want)
	service := account.NewService(memoryStore)

	for range 3 {
		if _, err := service.Balance("100"); err != nil {
			t.Fatalf("balance returned an unexpected error: %v", err)
		}
	}

	assertStoredAccount(t, memoryStore, want)
}

func TestResetRemovesAllAccounts(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	memoryStore.Save(account.Account{ID: "100", Balance: 10})
	memoryStore.Save(account.Account{ID: "200", Balance: 20})
	service := account.NewService(memoryStore)

	service.Reset()

	for _, id := range []string{"100", "200"} {
		if _, err := service.Balance(id); !errors.Is(err, account.ErrAccountNotFound) {
			t.Fatalf("expected account %q to be removed, got %v", id, err)
		}
	}
}

func TestResetAllowsNewStateAfterClearingAccounts(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	service := account.NewService(memoryStore)

	if _, err := service.Deposit("100", 10); err != nil {
		t.Fatalf("deposit returned an unexpected error: %v", err)
	}
	service.Reset()
	want := account.Account{ID: "100", Balance: 5}
	got, err := service.Deposit("100", 5)

	if err != nil {
		t.Fatalf("deposit after reset returned an unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected account %+v, got %+v", want, got)
	}
	assertStoredAccount(t, memoryStore, want)
}

func TestDepositCreatesDestinationAccount(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	service := account.NewService(memoryStore)
	want := account.Account{ID: "100", Balance: 10}

	got, err := service.Deposit("100", 10)

	if err != nil {
		t.Fatalf("deposit returned an unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected account %+v, got %+v", want, got)
	}
	assertStoredAccount(t, memoryStore, want)
}

func TestDepositCreditsExistingAccount(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	memoryStore.Save(account.Account{ID: "100", Balance: 10})
	service := account.NewService(memoryStore)
	want := account.Account{ID: "100", Balance: 20}

	got, err := service.Deposit("100", 10)

	if err != nil {
		t.Fatalf("deposit returned an unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected account %+v, got %+v", want, got)
	}
	assertStoredAccount(t, memoryStore, want)
}

func TestDepositRejectsInvalidAmountWithoutChangingState(t *testing.T) {
	testCases := []struct {
		name   string
		amount int64
	}{
		{name: "zero", amount: 0},
		{name: "negative", amount: -10},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			memoryStore := store.NewMemoryStore()
			original := account.Account{ID: "100", Balance: 10}
			memoryStore.Save(original)
			service := account.NewService(memoryStore)

			_, err := service.Deposit("100", testCase.amount)

			if !errors.Is(err, account.ErrInvalidAmount) {
				t.Fatalf("expected ErrInvalidAmount, got %v", err)
			}
			assertStoredAccount(t, memoryStore, original)
		})
	}
}

func TestDepositDoesNotCreateAccountForInvalidAmount(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	service := account.NewService(memoryStore)

	_, err := service.Deposit("100", 0)

	if !errors.Is(err, account.ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
	if _, found := memoryStore.Find("100"); found {
		t.Fatal("expected invalid deposit not to create an account")
	}
}

func TestConcurrentDepositsDoNotLoseUpdates(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	service := account.NewService(memoryStore)
	const depositCount = 100

	var waitGroup sync.WaitGroup
	for range depositCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if _, err := service.Deposit("100", 1); err != nil {
				t.Errorf("deposit returned an unexpected error: %v", err)
			}
		}()
	}
	waitGroup.Wait()

	assertStoredAccount(t, memoryStore, account.Account{ID: "100", Balance: depositCount})
}

func assertStoredAccount(t *testing.T, memoryStore *store.MemoryStore, want account.Account) {
	t.Helper()

	got, found := memoryStore.Find(want.ID)
	if !found {
		t.Fatalf("expected account %q to be stored", want.ID)
	}
	if got != want {
		t.Fatalf("expected stored account %+v, got %+v", want, got)
	}
}

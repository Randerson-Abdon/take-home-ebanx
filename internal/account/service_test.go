package account_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/store"
)

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

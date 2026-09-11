package account_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
	"github.com/Randerson-Abdon/take-home-ebanx/internal/store"
)

func TestTransferDebitsOriginAndCreditsDestination(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	memoryStore.Save(account.Account{ID: "100", Balance: 20})
	memoryStore.Save(account.Account{ID: "300", Balance: 5})
	service := account.NewService(memoryStore)
	wantOrigin := account.Account{ID: "100", Balance: 5}
	wantDestination := account.Account{ID: "300", Balance: 20}

	gotOrigin, gotDestination, err := service.Transfer("100", "300", 15)

	if err != nil {
		t.Fatalf("transfer returned an unexpected error: %v", err)
	}
	if gotOrigin != wantOrigin {
		t.Fatalf("expected origin %+v, got %+v", wantOrigin, gotOrigin)
	}
	if gotDestination != wantDestination {
		t.Fatalf("expected destination %+v, got %+v", wantDestination, gotDestination)
	}
	assertStoredAccount(t, memoryStore, wantOrigin)
	assertStoredAccount(t, memoryStore, wantDestination)
}

func TestTransferCreatesMissingDestination(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	memoryStore.Save(account.Account{ID: "100", Balance: 15})
	service := account.NewService(memoryStore)
	wantOrigin := account.Account{ID: "100", Balance: 0}
	wantDestination := account.Account{ID: "300", Balance: 15}

	gotOrigin, gotDestination, err := service.Transfer("100", "300", 15)

	if err != nil {
		t.Fatalf("transfer returned an unexpected error: %v", err)
	}
	if gotOrigin != wantOrigin {
		t.Fatalf("expected origin %+v, got %+v", wantOrigin, gotOrigin)
	}
	if gotDestination != wantDestination {
		t.Fatalf("expected destination %+v, got %+v", wantDestination, gotDestination)
	}
	assertStoredAccount(t, memoryStore, wantOrigin)
	assertStoredAccount(t, memoryStore, wantDestination)
}

func TestTransferReturnsErrorForMissingOriginWithoutChangingDestination(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	originalDestination := account.Account{ID: "300", Balance: 5}
	memoryStore.Save(originalDestination)
	service := account.NewService(memoryStore)

	_, _, err := service.Transfer("missing", "300", 10)

	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
	if _, found := memoryStore.Find("missing"); found {
		t.Fatal("expected transfer not to create the missing origin")
	}
	assertStoredAccount(t, memoryStore, originalDestination)
}

func TestTransferDoesNotCreateDestinationWhenOriginIsMissing(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	service := account.NewService(memoryStore)

	_, _, err := service.Transfer("missing", "300", 10)

	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
	if _, found := memoryStore.Find("300"); found {
		t.Fatal("expected failed transfer not to create the destination")
	}
}

func TestTransferRejectsInvalidAmountWithoutChangingState(t *testing.T) {
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
			originalOrigin := account.Account{ID: "100", Balance: 20}
			originalDestination := account.Account{ID: "300", Balance: 5}
			memoryStore.SaveAll(originalOrigin, originalDestination)
			service := account.NewService(memoryStore)

			_, _, err := service.Transfer("100", "300", testCase.amount)

			if !errors.Is(err, account.ErrInvalidAmount) {
				t.Fatalf("expected ErrInvalidAmount, got %v", err)
			}
			assertStoredAccount(t, memoryStore, originalOrigin)
			assertStoredAccount(t, memoryStore, originalDestination)
		})
	}
}

func TestTransferRejectsInsufficientFundsWithoutChangingState(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	originalOrigin := account.Account{ID: "100", Balance: 10}
	originalDestination := account.Account{ID: "300", Balance: 5}
	memoryStore.SaveAll(originalOrigin, originalDestination)
	service := account.NewService(memoryStore)

	_, _, err := service.Transfer("100", "300", 15)

	if !errors.Is(err, account.ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
	assertStoredAccount(t, memoryStore, originalOrigin)
	assertStoredAccount(t, memoryStore, originalDestination)
}

func TestTransferToSameAccountDoesNotChangeBalance(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	original := account.Account{ID: "100", Balance: 20}
	memoryStore.Save(original)
	service := account.NewService(memoryStore)

	gotOrigin, gotDestination, err := service.Transfer("100", "100", 10)

	if err != nil {
		t.Fatalf("transfer returned an unexpected error: %v", err)
	}
	if gotOrigin != original || gotDestination != original {
		t.Fatalf("expected both results to remain %+v, got %+v and %+v", original, gotOrigin, gotDestination)
	}
	assertStoredAccount(t, memoryStore, original)
}

func TestConcurrentTransfersDoNotOverdrawOrigin(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	memoryStore.Save(account.Account{ID: "100", Balance: 100})
	service := account.NewService(memoryStore)
	const transferCount = 150

	var successfulTransfers atomic.Int64
	var rejectedTransfers atomic.Int64
	var waitGroup sync.WaitGroup
	for range transferCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			_, _, err := service.Transfer("100", "300", 1)
			switch {
			case err == nil:
				successfulTransfers.Add(1)
			case errors.Is(err, account.ErrInsufficientFunds):
				rejectedTransfers.Add(1)
			default:
				t.Errorf("transfer returned an unexpected error: %v", err)
			}
		}()
	}
	waitGroup.Wait()

	if successfulTransfers.Load() != 100 {
		t.Fatalf("expected 100 successful transfers, got %d", successfulTransfers.Load())
	}
	if rejectedTransfers.Load() != 50 {
		t.Fatalf("expected 50 rejected transfers, got %d", rejectedTransfers.Load())
	}
	assertStoredAccount(t, memoryStore, account.Account{ID: "100", Balance: 0})
	assertStoredAccount(t, memoryStore, account.Account{ID: "300", Balance: 100})
}

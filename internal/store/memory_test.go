package store

import (
	"strconv"
	"sync"
	"testing"

	"github.com/Randerson-Abdon/take-home-ebanx/internal/account"
)

func TestMemoryStoreFindMissingAccount(t *testing.T) {
	store := NewMemoryStore()

	_, found := store.Find("missing")

	if found {
		t.Fatal("expected account not to be found")
	}
}

func TestMemoryStoreSaveAndFindAccount(t *testing.T) {
	store := NewMemoryStore()
	want := account.Account{ID: "100", Balance: 10}

	store.Save(want)
	got, found := store.Find(want.ID)

	if !found {
		t.Fatal("expected account to be found")
	}
	if got != want {
		t.Fatalf("expected account %+v, got %+v", want, got)
	}
}

func TestMemoryStoreUpdatesExistingAccount(t *testing.T) {
	store := NewMemoryStore()
	store.Save(account.Account{ID: "100", Balance: 10})
	want := account.Account{ID: "100", Balance: 20}

	store.Save(want)
	got, found := store.Find(want.ID)

	if !found {
		t.Fatal("expected account to be found")
	}
	if got != want {
		t.Fatalf("expected account %+v, got %+v", want, got)
	}
}

func TestMemoryStoreSavesMultipleAccounts(t *testing.T) {
	store := NewMemoryStore()
	accounts := []account.Account{
		{ID: "100", Balance: 5},
		{ID: "200", Balance: 15},
	}

	store.SaveAll(accounts...)

	for _, want := range accounts {
		got, found := store.Find(want.ID)
		if !found {
			t.Fatalf("expected account %q to be found", want.ID)
		}
		if got != want {
			t.Fatalf("expected account %+v, got %+v", want, got)
		}
	}
}

func TestMemoryStoreReturnsAccountByValue(t *testing.T) {
	store := NewMemoryStore()
	store.Save(account.Account{ID: "100", Balance: 10})

	got, _ := store.Find("100")
	got.Balance = 999
	storedAccount, _ := store.Find("100")

	if storedAccount.Balance != 10 {
		t.Fatalf("expected stored balance to remain 10, got %d", storedAccount.Balance)
	}
}

func TestMemoryStoreReset(t *testing.T) {
	store := NewMemoryStore()
	store.Save(account.Account{ID: "100", Balance: 10})
	store.Save(account.Account{ID: "200", Balance: 20})

	store.Reset()

	for _, id := range []string{"100", "200"} {
		if _, found := store.Find(id); found {
			t.Fatalf("expected account %q to be removed", id)
		}
	}
}

func TestMemoryStoreSupportsConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()
	const accountCount = 100

	var waitGroup sync.WaitGroup
	for index := 0; index < accountCount; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			id := strconv.Itoa(index)
			store.Save(account.Account{ID: id, Balance: int64(index)})
			store.Find(id)
		}()
	}
	waitGroup.Wait()

	for index := 0; index < accountCount; index++ {
		id := strconv.Itoa(index)
		got, found := store.Find(id)
		if !found {
			t.Fatalf("expected account %q to be found", id)
		}
		if got.Balance != int64(index) {
			t.Fatalf("expected balance %d for account %q, got %d", index, id, got.Balance)
		}
	}
}

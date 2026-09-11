// Package account contains the account domain model and business rules.
package account

// Account represents the current balance of an account.
type Account struct {
	ID      string
	Balance int64
}

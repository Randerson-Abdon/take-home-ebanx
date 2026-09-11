package account

import "errors"

var (
	// ErrAccountNotFound indicates that an account does not exist.
	ErrAccountNotFound = errors.New("account not found")

	// ErrInvalidAmount indicates that an operation received a non-positive amount.
	ErrInvalidAmount = errors.New("amount must be greater than zero")

	// ErrInsufficientFunds indicates that an account cannot cover a debit.
	ErrInsufficientFunds = errors.New("insufficient funds")
)

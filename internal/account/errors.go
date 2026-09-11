package account

import "errors"

// ErrInvalidAmount indicates that an operation received a non-positive amount.
var ErrInvalidAmount = errors.New("amount must be greater than zero")

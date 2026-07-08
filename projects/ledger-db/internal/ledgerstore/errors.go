package ledgerstore

import "errors"

var (
	ErrAmountGreaterThanZero         = errors.New("amount must be greater than 0")
	ErrExternalReferenceRequired     = errors.New("external reference must not be empty")
	ErrNoRowsFound                   = errors.New("no rows found")
	ErrToAccountNotFound             = errors.New("to account not found")
	ErrCashSettlementAccountNotFound = errors.New("Cash Settlement account not found")
	ErrIdempotencyConflict           = errors.New("idempotency conflict")
	ErrCurrencyMismatch              = errors.New("currency mismatch")
	ErrInsufficientFUnds             = errors.New("insufficient funds")
)

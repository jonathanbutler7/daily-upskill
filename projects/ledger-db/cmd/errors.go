package ledger

import "errors"

const (
	dbErrInsufficientFunds   = "insufficient funds"
	dbErrIdempotencyConflict = "idempotency key reused with different request"
	dbErrFromAccountNotFound = "from account not found"
	dbErrToAccountNotFound   = "to account not found"
	dbErrCurrencyMismatch    = "currency mismatch"
)

var (
	ErrFromAccountIDRequired       = errors.New("from account id is required")
	ErrToAccountIDRequired         = errors.New("to account id is required")
	ErrAmountGreaterThanZero       = errors.New("amount must be greater than 0")
	ErrIdempotencyKeyRequired      = errors.New("idempotency key is required")
	ErrTransferAmountRequired      = errors.New("transfer amount is required")
	ErrRailValueRequired           = errors.New("rail value is required")
	ErrExternalReferenceIdRequired = errors.New("externalReferenceID is required")

	ErrInsufficientFunds             = errors.New("insufficient funds")
	ErrIdempotencyConflict           = errors.New("idempotency conflict")
	ErrFromAccountNotFound           = errors.New("from account not found")
	ErrToAccountNotFound             = errors.New("to account not found")
	ErrCurrencyMismatch              = errors.New("currency mismatch")
	ErrExternalReferenceIdEmpty      = errors.New("external reference must not be empty")
	ErrCashSettlementAccountNotFound = errors.New("Cash Settlement account not found")
)

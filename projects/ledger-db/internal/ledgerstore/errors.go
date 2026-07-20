package ledgerstore

import "errors"

var (
	ErrAmountGreaterThanZero         = errors.New("amount must be greater than 0")
	ErrCashSettlementAccountNotFound = errors.New("Cash Settlement account not found")
	ErrCurrencyMismatch              = errors.New("currency mismatch")
	ErrExternalReferenceEmpty        = errors.New("external reference must not be empty")
	ErrExternalReferenceRequired     = errors.New("externalReference is required")
	ErrFromAccountIDRequired         = errors.New("from account id is required")
	ErrFromAccountNotFound           = errors.New("from account not found")
	ErrIdempotencyConflict           = errors.New("idempotency conflict")
	ErrIdempotencyKeyRequired        = errors.New("idempotency key is required")
	ErrInsufficientFunds             = errors.New("insufficient funds")
	ErrMustBeWithdrawalOrDeposit     = errors.New("direction must be withdrawal or deposit")
	ErrNoRowsFound                   = errors.New("no rows found")
	ErrRailValueRequired             = errors.New("rail value is required")
	ErrReversalAlreadyExists         = errors.New("reversal already exists")
	ErrToAccountIDRequired           = errors.New("to account id is required")
	ErrToAccountNotFound             = errors.New("to account not found")
	ErrTransactionIDRequired         = errors.New("transaction id is required")
	ErrTransactionNotBalanced        = errors.New("transaction is not balanced")
	ErrTransferAmountRequired        = errors.New("transfer amount is required")
)

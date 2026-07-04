package ledger

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

const (
	dbErrInsufficientFunds   = "insufficient funds"
	dbErrIdempotencyConflict = "idempotency key reused with different request"
	dbErrFromAccountNotFound = "from account not found"
	dbErrToAccountNotFound   = "to account not found"
	dbErrCurrencyMismatch    = "currency mismatch"
)

var (
	ErrFromAccountIDRequired  = errors.New("from account id is required")
	ErrToAccountIDRequired    = errors.New("to account id is required")
	ErrAmountGreaterThanZero  = errors.New("amount must be greater than 0")
	ErrIdempotencyKeyRequired = errors.New("idempotency key is required")

	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrFromAccountNotFound = errors.New("from account not found")
	ErrToAccountNotFound   = errors.New("to account not found")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
)

func PostTransfer(
	ctx context.Context,
	db *sql.DB,
	fromAccountID,
	toAccountID,
	amount int64,
	idempotencyKey string,
) (int64, error) {
	if fromAccountID <= 0 {
		return 0, ErrFromAccountIDRequired
	}
	if toAccountID <= 0 {
		return 0, ErrToAccountIDRequired
	}
	if amount <= 0 {
		return 0, ErrAmountGreaterThanZero
	}
	if idempotencyKey == "" {
		return 0, ErrIdempotencyKeyRequired
	}

	var transactionID int64
	err := db.QueryRowContext(
		ctx,
		`select post_transfer($1, $2, $3, $4)`,
		fromAccountID, toAccountID, amount, idempotencyKey,
	).Scan(&transactionID)
	if err != nil && strings.Contains(err.Error(), dbErrInsufficientFunds) {
		return 0, ErrInsufficientFunds
	}
	if err != nil && strings.Contains(err.Error(), dbErrIdempotencyConflict) {
		return 0, ErrIdempotencyConflict
	}
	if err != nil && strings.Contains(err.Error(), dbErrFromAccountNotFound) {
		return 0, ErrFromAccountNotFound
	}
	if err != nil && strings.Contains(err.Error(), dbErrToAccountNotFound) {
		return 0, ErrToAccountNotFound
	}
	if err != nil && strings.Contains(err.Error(), dbErrCurrencyMismatch) {
		return 0, ErrCurrencyMismatch
	}

	return transactionID, nil
}

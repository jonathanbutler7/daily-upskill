package ledger

import (
	"context"
	"database/sql"
	"strings"
)

type TransferCommand struct {
	FromAccountID  int64
	ToAccountID    int64
	Amount         int64
	IdempotencyKey string
}

func PostTransfer(
	ctx context.Context,
	db *sql.DB,
	cmd TransferCommand,
) (int64, error) {
	if cmd.FromAccountID <= 0 {
		return 0, ErrFromAccountIDRequired
	}
	if cmd.ToAccountID <= 0 {
		return 0, ErrToAccountIDRequired
	}
	if cmd.Amount <= 0 {
		return 0, ErrAmountGreaterThanZero
	}
	if cmd.IdempotencyKey == "" {
		return 0, ErrIdempotencyKeyRequired
	}

	var transactionID int64
	err := db.QueryRowContext(
		ctx,
		`select post_transfer($1, $2, $3, $4)`,
		cmd.FromAccountID, cmd.ToAccountID, cmd.Amount, cmd.IdempotencyKey,
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
	// fallback to return unknown errors
	if err != nil {
		return 0, err
	}

	return transactionID, nil
}

package ledger

import (
	"context"
	"database/sql"
	"errors"
	"ledger-db/internal/ledgerstore"
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

	transactionID, err := ledgerstore.PostTransfer(
		ctx,
		db,
		ledgerstore.PostTransferCommand{
			FromAccountID:  ledgerstore.AccountID(cmd.FromAccountID),
			ToAccountID:    ledgerstore.AccountID(cmd.ToAccountID),
			Amount:         ledgerstore.Amount(cmd.Amount),
			IdempotencyKey: ledgerstore.IdempotencyKey(cmd.IdempotencyKey),
		},
	)
	if errors.Is(err, ErrInsufficientFunds) {
		return 0, ErrInsufficientFunds
	}
	if errors.Is(err, ErrIdempotencyConflict) {
		return 0, ErrIdempotencyConflict
	}
	if errors.Is(err, ErrFromAccountNotFound) {
		return 0, ErrFromAccountNotFound
	}
	if errors.Is(err, ErrToAccountNotFound) {
		return 0, ErrToAccountNotFound
	}
	if errors.Is(err, ErrCurrencyMismatch) {
		return 0, ErrCurrencyMismatch
	}
	// fallback to return unknown errors
	if err != nil {
		return 0, err
	}

	return int64(transactionID), nil
}

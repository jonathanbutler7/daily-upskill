package ledger

import (
	"context"
	"database/sql"
	"errors"
	"ledger-db/internal/ledgerstore"
)


func PostTransfer(
	ctx context.Context,
	db *sql.DB,
	cmd ledgerstore.TransferCommand,
) (int64, error) {
	if cmd.FromAccountID <= 0 {
		return 0, ledgerstore.ErrFromAccountIDRequired
	}
	if cmd.ToAccountID <= 0 {
		return 0, ledgerstore.ErrToAccountIDRequired
	}
	if cmd.Amount <= 0 {
		return 0, ledgerstore.ErrAmountGreaterThanZero
	}
	if cmd.IdempotencyKey == "" {
		return 0, ledgerstore.ErrIdempotencyKeyRequired
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
	if errors.Is(err, ledgerstore.ErrInsufficientFunds) {
		return 0, ledgerstore.ErrInsufficientFunds
	}
	if errors.Is(err, ledgerstore.ErrIdempotencyConflict) {
		return 0, ledgerstore.ErrIdempotencyConflict
	}
	if errors.Is(err, ledgerstore.ErrFromAccountNotFound) {
		return 0, ledgerstore.ErrFromAccountNotFound
	}
	if errors.Is(err, ledgerstore.ErrToAccountNotFound) {
		return 0, ledgerstore.ErrToAccountNotFound
	}
	if errors.Is(err, ledgerstore.ErrCurrencyMismatch) {
		return 0, ledgerstore.ErrCurrencyMismatch
	}
	// fallback to return unknown errors
	if err != nil {
		return 0, err
	}

	return int64(transactionID), nil
}

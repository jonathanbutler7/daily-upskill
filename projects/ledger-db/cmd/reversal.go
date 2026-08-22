package ledger

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ledger-db/internal/ledgerstore"
)

func Reversal(
	ctx context.Context,
	db *sql.DB,
	cmd ledgerstore.ReversalCommand,
) (int64, error) {
	if cmd.TransactionID <= 0 {
		return 0, ledgerstore.ErrTransactionIDRequired
	}
	if strings.TrimSpace(string(cmd.Reason)) == "" {
		return 0, ledgerstore.ErrReasonIsRequired
	}
	if cmd.IdempotencyKey == "" {
		return 0, ledgerstore.ErrIdempotencyKeyRequired
	}

	transactionID, err := ledgerstore.ReverseTransaction(
		ctx,
		db,
		ledgerstore.ReversalCommand{
			TransactionID:  ledgerstore.TransactionID(cmd.TransactionID),
			IdempotencyKey: ledgerstore.IdempotencyKey(cmd.IdempotencyKey),
			Reason:         ledgerstore.Reason(strings.TrimSpace(string(cmd.Reason))),
		},
	)
	if errors.Is(err, ledgerstore.ErrInsufficientFunds) {
		return 0, ledgerstore.ErrInsufficientFunds
	}
	if errors.Is(err, ledgerstore.ErrReversalAlreadyExists) {
		return 0, ledgerstore.ErrReversalAlreadyExists
	}
	if errors.Is(err, ledgerstore.ErrIdempotencyConflict) {
		return 0, ledgerstore.ErrIdempotencyConflict
	}
	if errors.Is(err, ledgerstore.ErrNoRowsFound) {
		return 0, ledgerstore.ErrNoRowsFound
	}
	if err != nil {
		return 0, err
	}

	return int64(transactionID), nil
}

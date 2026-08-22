package ledger

import (
	"context"
	"database/sql"
	"ledger-db/internal/ledgerstore"
)

func Reversal(ctx context.Context,
	db *sql.DB,
	cmd ledgerstore.ReversalCommand,
) (int64, error) {
	if cmd.TransactionID == 0 {
		return 0, ledgerstore.ErrTransactionIDRequired
	}
	if cmd.Reason == "" {
		return 0,ledgerstore.ErrReasonIsRequired
	}
	if cmd.IdempotencyKey == "" {
		return 0, ledgerstore.ErrIdempotencyKeyRequired
	}
	
	transactionID, err := ledgerstore.ReverseTransaction(ctx, db, cmd)
	if err != nil {
		return 0, err
	}
	return int64(transactionID), nil
}

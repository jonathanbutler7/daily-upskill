package ledger

import (
	"context"
	"database/sql"
	"fmt"
)

func AddBalance(
	ctx context.Context,
	db *sql.DB,
	toAccountID, transferAmount int64,
	rail, externalReferenceID, idempotencyKey string,
) (int64, error) {
	if toAccountID <= 0 {
		return 0, fmt.Errorf("to account id is required")
	}
	if transferAmount <= 0 {
		return 0, fmt.Errorf("transfer amount is required")
	}
	if rail == "" {
		return 0, fmt.Errorf("rail value is required")
	}
	if externalReferenceID == "" {
		return 0, fmt.Errorf("externalReferenceID is required")
	}
	if idempotencyKey == "" {
		return 0, fmt.Errorf("idempotencyKey is required")
	}

	var transactionID int64
	err := db.QueryRowContext(
		ctx,
		`select add_balance($1, $2, $3, $4, $5)`,
		toAccountID, transferAmount, rail, externalReferenceID, idempotencyKey,
	).Scan(&transactionID)
	
	return transactionID, err
}

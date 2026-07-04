package ledger

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	dbErrAmountGreaterThanZero         = "amount must be greater than 0"
	dbErrExternalReferenceIdEmpty      = "external reference must not be empty"
	dbErrCashSettlementAccountNotFound = "Cash Settlement account not found"
)

var (
	// request validation errors
	ErrTransferAmountRequired      = errors.New("transfer amount is required")
	ErrRailValueRequired           = errors.New("rail value is required")
	ErrExternalReferenceIdRequired = errors.New("externalReferenceID is required")

	// db errors
	ErrExternalReferenceIdEmpty      = errors.New(dbErrExternalReferenceIdEmpty)
	ErrCashSettlementAccountNotFound = errors.New(dbErrCashSettlementAccountNotFound)
)

type AddBalanceCommand struct {
	ToAccountID         int64
	TransferAmount      int64
	Rail                string
	ExternalReferenceID string
	IdempotencyKey      string
}

func AddBalance(
	ctx context.Context,
	db *sql.DB,
	cmd AddBalanceCommand,
) (int64, error) {
	if cmd.ToAccountID <= 0 {
		return 0, ErrToAccountIDRequired
	}
	if cmd.TransferAmount <= 0 {
		return 0, ErrTransferAmountRequired
	}
	if cmd.Rail == "" {
		return 0, ErrRailValueRequired
	}
	if cmd.ExternalReferenceID == "" {
		return 0, ErrExternalReferenceIdRequired
	}
	if cmd.IdempotencyKey == "" {
		return 0, ErrIdempotencyKeyRequired
	}

	var transactionID int64
	err := db.QueryRowContext(
		ctx,
		`select add_balance($1, $2, $3, $4, $5)`,
		cmd.ToAccountID, cmd.TransferAmount, cmd.Rail, cmd.ExternalReferenceID, cmd.IdempotencyKey,
	).Scan(&transactionID)

	if err != nil && strings.Contains(err.Error(), dbErrAmountGreaterThanZero) {
		return 0, ErrAmountGreaterThanZero
	}
	if err != nil && strings.Contains(err.Error(), dbErrExternalReferenceIdEmpty) {
		return 0, ErrExternalReferenceIdEmpty
	}
	if err != nil && strings.Contains(err.Error(), dbErrToAccountNotFound) {
		return 0, ErrToAccountNotFound
	}
	if err != nil && strings.Contains(err.Error(), dbErrCashSettlementAccountNotFound) {
		return 0, ErrCashSettlementAccountNotFound
	}
	if err != nil && strings.Contains(err.Error(), dbErrIdempotencyConflict) {
		return 0, ErrIdempotencyConflict
	}
	// fallback to return unknown errors
	if err != nil {
		return 0, err
	}

	return transactionID, nil
}

package ledger

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ledger-db/internal/ledgerstore"
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

type DepositFundsCommand struct {
	ToAccountID         int64
	TransferAmount      int64
	Rail                string
	ExternalReferenceID string
	IdempotencyKey      string
}

func DepositFunds(
	ctx context.Context,
	db *sql.DB,
	cmd DepositFundsCommand,
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
	if strings.TrimSpace(cmd.ExternalReferenceID) == "" {
		return 0, ErrExternalReferenceIdRequired
	}
	if cmd.IdempotencyKey == "" {
		return 0, ErrIdempotencyKeyRequired
	}

	transactionID, err := ledgerstore.AddDeposit(
		ctx,
		db,
		ledgerstore.AddDepositCommand{
			ToAccountID:       ledgerstore.AccountID(cmd.ToAccountID),
			TransferAmount:    ledgerstore.Amount(cmd.TransferAmount),
			Rail:              ledgerstore.PaymentRail(cmd.Rail),
			ExternalReference: ledgerstore.ExternalReference(cmd.ExternalReferenceID),
			IdempotencyKey:    ledgerstore.IdempotencyKey(cmd.IdempotencyKey),
		})

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
	if errors.Is(err, ledgerstore.ErrIdempotencyConflict) {
		return 0, ErrIdempotencyConflict
	}
	// fallback to return unknown errors
	if err != nil {
		return 0, err
	}

	return int64(transactionID), nil
}

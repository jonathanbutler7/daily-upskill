package ledger

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ledger-db/internal/ledgerstore"
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
) (ledgerstore.TransactionID, error) {
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

	if errors.Is(err, ledgerstore.ErrAmountGreaterThanZero) {
		return 0, ErrAmountGreaterThanZero
	}
	if errors.Is(err, ledgerstore.ErrExternalReferenceRequired) {
		return 0, ErrExternalReferenceIdEmpty
	}
	if errors.Is(err, ledgerstore.ErrToAccountNotFound) {
		return 0, ErrToAccountNotFound
	}
	if errors.Is(err, ledgerstore.ErrCashSettlementAccountNotFound) {
		return 0, ErrCashSettlementAccountNotFound
	}
	if errors.Is(err, ledgerstore.ErrIdempotencyConflict) {
		return 0, ErrIdempotencyConflict
	}
	// fallback to return unknown errors
	if err != nil {
		return 0, err
	}

	return ledgerstore.TransactionID(transactionID), nil
}

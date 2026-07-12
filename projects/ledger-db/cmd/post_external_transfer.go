package ledger

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ledger-db/internal/ledgerstore"
)

func PostExternalTransfer(
	ctx context.Context,
	db *sql.DB,
	cmd ledgerstore.PostExternalTransferCommand,
) (ledgerstore.TransactionID, error) {
	if cmd.TransferAmount <= 0 {
		return 0, ledgerstore.ErrTransferAmountRequired
	}
	if cmd.ToAccountID <= 0 {
		return 0, ledgerstore.ErrToAccountIDRequired
	}
	if cmd.Rail == "" {
		return 0, ledgerstore.ErrRailValueRequired
	}
	if strings.TrimSpace(string(cmd.ExternalReference)) == "" {
		return 0, ledgerstore.ErrExternalReferenceRequired
	}
	if cmd.IdempotencyKey == "" {
		return 0, ledgerstore.ErrIdempotencyKeyRequired
	}

	transactionID, err := ledgerstore.PostExternalTransfer(
		ctx,
		db,
		ledgerstore.PostExternalTransferCommand{
			ToAccountID:       ledgerstore.AccountID(cmd.ToAccountID),
			TransferAmount:    ledgerstore.Amount(cmd.TransferAmount),
			Rail:              ledgerstore.PaymentRail(cmd.Rail),
			ExternalReference: ledgerstore.ExternalReference(cmd.ExternalReference),
			IdempotencyKey:    ledgerstore.IdempotencyKey(cmd.IdempotencyKey),
		})

	if errors.Is(err, ledgerstore.ErrAmountGreaterThanZero) {
		return 0, ledgerstore.ErrAmountGreaterThanZero
	}
	if errors.Is(err, ledgerstore.ErrExternalReferenceRequired) {
		return 0, ledgerstore.ErrExternalReferenceEmpty
	}
	if errors.Is(err, ledgerstore.ErrToAccountNotFound) {
		return 0, ledgerstore.ErrToAccountNotFound
	}
	if errors.Is(err, ledgerstore.ErrCashSettlementAccountNotFound) {
		return 0, ledgerstore.ErrCashSettlementAccountNotFound
	}
	if errors.Is(err, ledgerstore.ErrIdempotencyConflict) {
		return 0, ledgerstore.ErrIdempotencyConflict
	}
	// fallback to return unknown errors
	if err != nil {
		return 0, err
	}

	return ledgerstore.TransactionID(transactionID), nil
}

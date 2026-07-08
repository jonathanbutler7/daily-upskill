package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type AddDepositCommand struct {
	ToAccountID       AccountID
	TransferAmount    Amount
	Rail              PaymentRail
	ExternalReference ExternalReference
	IdempotencyKey    IdempotencyKey
}

func AddDeposit(ctx context.Context, db *sql.DB, cmd AddDepositCommand) (TransactionID, error) {
	if cmd.TransferAmount <= 0 {
		return 0, ErrAmountGreaterThanZero
	}
	if strings.TrimSpace(string(cmd.ExternalReference)) == "" {
		return 0, ErrExternalReferenceIdRequired
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	toAccountCurrency, err := lockToAccountCurrencyForUpdate(ctx, tx, cmd.ToAccountID)
	if err != nil {
		return 0, err
	}

	fundingAccountID, err := lockCashSettlementAccountForUpdate(ctx, tx, toAccountCurrency)
	if err != nil {
		return 0, err
	}

	transactionID, err := checkSameIdempotencyRequest(ctx, tx,
		cmd.IdempotencyKey,
		fundingAccountID,
		cmd.ToAccountID,
		cmd.TransferAmount,
		toAccountCurrency,
	)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if transactionID != 0 {
		return TransactionID(transactionID), nil
	}

	conflictingTransactionID, err := checkIdempotencyConflict(ctx, tx, cmd.IdempotencyKey)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if conflictingTransactionID != 0 {
		return 0, ErrIdempotencyConflict
	}

	transactionID, err = insertLedgerTransaction(
		ctx,
		tx,
		cmd.IdempotencyKey,
		fundingAccountID,
		cmd.ToAccountID,
		cmd.TransferAmount,
		toAccountCurrency,
	)
	if err != nil {
		return 0, err
	}

	if err := insertLedgerEntries(ctx, tx, transactionID, cmd.TransferAmount, fundingAccountID, cmd.ToAccountID); err != nil {
		return 0, err
	}

	if err := updateBalances(ctx, tx, cmd.TransferAmount, fundingAccountID, cmd.ToAccountID); err != nil {
		return 0, err
	}

	if err := insertExternalTransfers(
		ctx,
		tx,
		ExternalTransferDirection("deposit"),
		cmd.Rail,
		ExternalTransferStatus("posted"),
		cmd.ExternalReference,
		cmd.ToAccountID,
		transactionID,
		cmd.TransferAmount,
		toAccountCurrency,
	); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return TransactionID(transactionID), nil
}

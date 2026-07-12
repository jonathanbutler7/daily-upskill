package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func PostExternalTransfer(ctx context.Context, db *sql.DB, cmd PostExternalTransferCommand) (TransactionID, error) {
	if cmd.TransferAmount <= 0 {
		return 0, ErrAmountGreaterThanZero
	}
	if strings.TrimSpace(string(cmd.ExternalReference)) == "" {
		return 0, ErrExternalReferenceRequired
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

	transactionID, err := findSameLedgerTransaction(ctx, tx,
		"deposit",
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

	conflictingTransactionID, err := findLedgerTransactionByIdempotencyKey(ctx, tx, cmd.IdempotencyKey)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if conflictingTransactionID != 0 {
		return 0, ErrIdempotencyConflict
	}

	transactionID, err = insertLedgerTransaction(
		ctx,
		tx,
		"deposit",
		cmd.IdempotencyKey,
		fundingAccountID,
		cmd.ToAccountID,
		cmd.TransferAmount,
		toAccountCurrency,
	)
	if errors.Is(err, ErrNoRowsFound) {
		transactionID, err = findSameLedgerTransaction(
			ctx,
			tx,
			"deposit",
			cmd.IdempotencyKey,
			fundingAccountID,
			cmd.ToAccountID,
			cmd.TransferAmount,
			toAccountCurrency,
		)
		if err == nil {
			return transactionID, nil
		}
		if errors.Is(err, ErrNoRowsFound) {
			return 0, ErrIdempotencyConflict
		}
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	if err := insertLedgerEntries(ctx, tx, transactionID, cmd.TransferAmount, fundingAccountID, cmd.ToAccountID); err != nil {
		return 0, err
	}
	err = verifyTransactionBalances(ctx, tx, transactionID)
	if err != nil {
		return 0, err
	}

	if err := adjustAccountBalance(ctx, tx, fundingAccountID, -cmd.TransferAmount); err != nil {
		return 0, err
	}
	if err := adjustAccountBalance(ctx, tx, cmd.ToAccountID, cmd.TransferAmount); err != nil {
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

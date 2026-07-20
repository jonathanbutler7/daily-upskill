package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
)

func PostTransfer(ctx context.Context, db *sql.DB, cmd PostTransferCommand) (TransactionID, error) {
	if cmd.Amount <= 0 {
		return 0, ErrAmountGreaterThanZero
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	fromBalance, fromCurrency, err := lockAccountForUpdate(ctx, tx, cmd.FromAccountID)
	if err != nil {
		return 0, err
	}

	toCurrency, err := lockToAccountCurrencyForUpdate(ctx, tx, cmd.ToAccountID)
	if err != nil {
		return 0, err
	}

	if err := checkCurrencyMatch(fromCurrency, toCurrency); err != nil {
		return 0, err
	}

	transactionID, err := findSameLedgerTransaction(
		ctx,
		tx,
		LedgerTransactionTypeTransfer,
		cmd.IdempotencyKey,
		cmd.FromAccountID,
		cmd.ToAccountID,
		cmd.Amount,
		fromCurrency,
	)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if transactionID != 0 {
		return transactionID, nil
	}

	conflictingTransactionID, err := findLedgerTransactionByIdempotencyKey(
		ctx,
		tx,
		cmd.IdempotencyKey,
	)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if conflictingTransactionID != 0 {
		return 0, ErrIdempotencyConflict
	}

	if err := checkBalance(fromBalance, cmd.Amount); err != nil {
		return 0, err
	}

	transactionID, err = insertLedgerTransaction(
		ctx,
		tx,
		LedgerTransactionTypeTransfer,
		cmd.IdempotencyKey,
		cmd.FromAccountID,
		cmd.ToAccountID,
		cmd.Amount,
		fromCurrency,
	)
	if errors.Is(err, ErrNoRowsFound) {
		transactionID, err = findSameLedgerTransaction(
			ctx,
			tx,
			LedgerTransactionTypeTransfer,
			cmd.IdempotencyKey,
			cmd.FromAccountID,
			cmd.ToAccountID,
			cmd.Amount,
			fromCurrency,
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

	if err := insertLedgerEntries(ctx, tx, transactionID, cmd.Amount, cmd.FromAccountID, cmd.ToAccountID); err != nil {
		return 0, err
	}

	err = verifyTransactionBalances(ctx, tx, transactionID)
	if err != nil {
		return 0, err
	}

	if err := adjustAccountBalance(ctx, tx, cmd.FromAccountID, -cmd.Amount); err != nil {
		return 0, err
	}
	if err := adjustAccountBalance(ctx, tx, cmd.ToAccountID, cmd.Amount); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return transactionID, nil
}

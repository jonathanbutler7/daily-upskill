package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
)

type PostTransferCommand struct {
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	IdempotencyKey IdempotencyKey
}

func PostTransfer(ctx context.Context, db *sql.DB, cmd PostTransferCommand) (TransactionID, error) {
	if cmd.Amount <= 0 {
		return 0, ErrAmountGreaterThanZero
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	fromBalance, fromCurrency, err := lockFromAccount(ctx, tx, cmd.FromAccountID)
	if err != nil {
		return 0, err
	}

	toCurrency, err := lockToAccount(ctx, tx, cmd.ToAccountID)
	if err != nil {
		return 0, err
	}

	if err := checkCurrencyMatch(fromCurrency, toCurrency); err != nil {
		return 0, err
	}

	transactionID, err := checkIdempotencyRequest(
		ctx,
		tx,
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

	conflictingTransactionID, err := checkIdempotencyConflict2(
		ctx,
		tx,
		cmd.IdempotencyKey,
		cmd.FromAccountID,
		cmd.ToAccountID,
		cmd.Amount,
		fromCurrency,
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

	transactionID, err = insertTransaction(
		ctx,
		tx,
		cmd.IdempotencyKey,
		cmd.FromAccountID,
		cmd.ToAccountID,
		cmd.Amount,
		fromCurrency,
	)
	if err != nil {
		return 0, err
	}

	if err := insertEntries(ctx, tx, transactionID, cmd.Amount, cmd.FromAccountID, cmd.ToAccountID); err != nil {
		return 0, err
	}

	if err := updateTransferBalances(ctx, tx, cmd.Amount, cmd.FromAccountID, cmd.ToAccountID); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return transactionID, nil
}

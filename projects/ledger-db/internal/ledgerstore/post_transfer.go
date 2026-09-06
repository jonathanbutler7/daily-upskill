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
	if cmd.IdempotencyKey == "" {
		return 0, ErrIdempotencyKeyRequired
	}
	if cmd.FromAccountID == 0 {
		return 0, ErrFromAccountIDRequired
	}
	if cmd.ToAccountID == 0 {
		return 0, ErrToAccountIDRequired
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	fromAccount, err := getAccountById(ctx, tx, cmd.FromAccountID)
	if errors.Is(err, ErrNoRowsFound) {
		return 0, ErrFromAccountNotFound
	}
	if err != nil {
		return 0, err
	}

	toAccount, err := getAccountById(ctx, tx, cmd.ToAccountID)
	if errors.Is(err, ErrNoRowsFound) {
		return 0, ErrToAccountNotFound
	}
	if err != nil {
		return 0, err
	}

	if err := checkCurrencyMatch(fromAccount.CurrencyCode, toAccount.CurrencyCode); err != nil {
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
		fromAccount.CurrencyCode,
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

	transactionID, err = insertLedgerTransaction(
		ctx,
		tx,
		LedgerTransactionTypeTransfer,
		cmd.IdempotencyKey,
		cmd.FromAccountID,
		cmd.ToAccountID,
		cmd.Amount,
		fromAccount.CurrencyCode,
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
			fromAccount.CurrencyCode,
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

	entries := []LedgerEntryInput{
		{AccountID: cmd.FromAccountID, Amount: cmd.Amount, Direction: EntryDirectionDebit},
		{AccountID: cmd.ToAccountID, Amount: cmd.Amount, Direction: EntryDirectionCredit},
	}
	for _, entry := range entries {
		if err := insertLedgerEntry(ctx, tx, transactionID, entry); err != nil {
			return 0, err
		}
	}

	err = verifyTransactionBalances(ctx, tx, transactionID)
	if err != nil {
		return 0, err
	}

	if err := adjustAccountBalance(ctx, tx, cmd.FromAccountID, cmd.Amount, EntryDirectionDebit); err != nil {
		return 0, err
	}
	if err := adjustAccountBalance(ctx, tx, cmd.ToAccountID, cmd.Amount, EntryDirectionCredit); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return transactionID, nil
}

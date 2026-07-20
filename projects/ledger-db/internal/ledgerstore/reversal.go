package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
)

func ReverseTransaction(ctx context.Context, db *sql.DB, cmd ReversalCommand) (TransactionID, error) {
	if cmd.TransactionID == 0 {
		return 0, ErrTransactionIDRequired
	}
	if cmd.IdempotencyKey == "" {
		return 0, ErrIdempotencyKeyRequired
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	transaction, err := findAndLockOriginalTransaction(ctx, tx, cmd.TransactionID)
	if err != nil {
		return 0, err
	}

	existingReversalID, err := findReversalByOriginalTransactionId(ctx, tx, cmd.TransactionID)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if existingReversalID != 0 {
		return 0, ErrReversalAlreadyExists
	}

	entries, err := getEntriesByTransactionId(ctx, tx, transaction.ID)
	if err != nil {
		return 0, err
	}

	reversalEntries := make([]LedgerEntryInput, 0, len(entries))
	for _, entry := range entries {
		reversalEntries = append(reversalEntries, LedgerEntryInput{
			AccountID: entry.AccountID,
			Amount:    -entry.Amount,
		})
	}

	reversalTransactionID, err := insertLedgerTransaction(
		ctx,
		tx,
		LedgerTransactionTypeReversal,
		cmd.IdempotencyKey,
		transaction.ToAccountID,
		transaction.FromAccountID,
		transaction.Amount,
		transaction.CurrencyCode,
	)
	if err != nil {
		return 0, err
	}

	for _, entry := range reversalEntries {
		currentBalance, _, err := lockAccountForUpdate(ctx, tx, entry.AccountID)
		if err != nil {
			return 0, err
		}
		if entry.Amount < 0 {
			if err := checkBalance(currentBalance, -entry.Amount); err != nil {
				return 0, err
			}
		}
		if err := insertLedgerEntry(ctx, tx, reversalTransactionID, entry); err != nil {
			return 0, err
		}
	}

	if err := verifyTransactionBalances(ctx, tx, reversalTransactionID); err != nil {
		return 0, err
	}

	for _, entry := range reversalEntries {
		if err := adjustAccountBalance(ctx, tx, entry.AccountID, entry.Amount); err != nil {
			return 0, err
		}
	}

	if err := insertLedgerReversal(ctx, tx, cmd.TransactionID, reversalTransactionID, cmd.Reason); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return reversalTransactionID, nil
}

func insertLedgerReversal(ctx context.Context, tx *sql.Tx, originalTransactionID TransactionID, reversalTransactionID TransactionID, reason Reason) error {
	const q = `
		insert into ledger_reversals
			(original_transaction_id, reversal_transaction_id, reason)
		values
			($1, $2, $3);
	`

	_, err := tx.ExecContext(ctx, q, originalTransactionID, reversalTransactionID, reason)
	return err
}

func findAndLockOriginalTransaction(ctx context.Context, tx *sql.Tx, transactionID TransactionID) (Transaction, error) {
	const q = `
		select id, type, idempotency_key, created_at::text, from_account_id, to_account_id, amount, currency_code
		from ledger_transactions
		where id = $1
		for update;
	`

	var transaction Transaction
	err := tx.QueryRowContext(ctx, q, transactionID).Scan(
		&transaction.ID,
		&transaction.Type,
		&transaction.IdempotencyKey,
		&transaction.CreatedAt,
		&transaction.FromAccountID,
		&transaction.ToAccountID,
		&transaction.Amount,
		&transaction.CurrencyCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, ErrNoRowsFound
	}
	if err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func findReversalByOriginalTransactionId(ctx context.Context, tx *sql.Tx, transactionID TransactionID) (TransactionID, error) {
	const q = `
		select reversal_transaction_id
		from ledger_reversals
		where original_transaction_id = $1;
	`

	var reversalTransactionID TransactionID
	err := tx.QueryRowContext(ctx, q, transactionID).Scan(&reversalTransactionID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoRowsFound
	}
	if err != nil {
		return 0, err
	}
	return reversalTransactionID, nil
}

func getEntriesByTransactionId(ctx context.Context, tx *sql.Tx, transactionID TransactionID) ([]Entry, error) {
	const q = `
		select id, transaction_id, account_id, amount, created_at::text
		from ledger_entries
		where transaction_id = $1
		order by account_id, id;
	`

	rows, err := tx.QueryContext(ctx, q, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.Amount,
			&entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

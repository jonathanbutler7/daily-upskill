package ledgerstore

import (
	"context"
	"database/sql"
)

func ReverseTransaction(ctx context.Context, db *sql.DB, cmd ReversalCommand) (TransactionID, error) {
	if cmd.TransactionID == 0 {
		return 0, ErrTransactionIDRequired
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// find and lock original transaction
	transaction, err := findAndLockOriginalTransaction(ctx, tx, cmd.TransactionID)
	if err != nil {
		return 0, err
	}

	// does reversal transaction already exist in reversals table?
	originalTransactionId, err := findReversalByOriginalTransactionId(ctx, tx, cmd.TransactionID)
	if originalTransactionId != 0 {
		return 0, ErrReversalAlreadyExists
	}

	// get entries for existing transaction
	entries, err := getEntriesByTransactionId(ctx, tx, transaction.ID)
	if err != nil {
		return 0, err
	}
	var fromAccountID AccountID
	var toAccountID AccountID

	for _, entry := range entries {
		if entry.AccountID == transaction.FromAccountID {
			fromAccountID = transaction.FromAccountID
		}
		if entry.AccountID == transaction.ToAccountID {
			toAccountID = transaction.ToAccountID
		}
	}

	// create a new reversal transaction in transactions type = 'reversal'
	reversalTransactionID, err := insertLedgerTransaction(
		ctx,
		tx,
		LedgerTransactionTypeReversal,
		"123",
		transaction.FromAccountID,
		transaction.ToAccountID,
		transaction.Amount,
		transaction.CurrencyCode,
	)
	// create opposite entries for each of the found entries
	err = insertLedgerEntries(
		ctx,
		tx,
		reversalTransactionID,
		transaction.Amount,
		fromAccountID,
		toAccountID,
	)

	// verify and update account balances
	currentFromAccountAmount, _, err := lockAccountForUpdate(ctx, tx, transaction.FromAccountID)
	if err != nil {
		return 0, err
	}
	currentToAccountAmount, _, err := lockAccountForUpdate(ctx, tx, transaction.ToAccountID)
	if err != nil {
		return 0, err
	}
	if err := adjustAccountBalance(ctx, tx, transaction.FromAccountID, -transaction.Amount); err != nil {
		return 0, err
	}
	if err := adjustAccountBalance(ctx, tx, transaction.ToAccountID, -transaction.Amount); err != nil {
		return 0, err
	}

	// create new entry in reversals
	// build this
	reversalId, err := insertLedgerReversal(ctx, tx, cmd.TransactionID, reversalTransactionID, cmd.Reason)

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return reversalId, nil
}
func insertLedgerReversal(ctx context.Context, tx *sql.Tx, originalTransactionId TransactionID, reversalTransactionId TransactionID, reason Reason) (TransactionID, error) {
	const q = `
		insert into ledger_reversals
		(original_transaction_id, reversal_transaction_id, reason, created_at)
		values ($1, $2, $3, $4);
	`
	var reversalId TransactionID

	return reversalId, nil
}

type Transaction struct {
	ID             TransactionID
	Type           LedgerTransactionType
	IdempotencyKey IdempotencyKey
	CreatedAt      string
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	CurrencyCode   CurrencyCode
}

type Entry struct {
	ID            string
	TransactionID TransactionID
	AccountID     AccountID
	Amount        Amount
	CreatedAt     string
}

func findAndLockOriginalTransaction(ctx context.Context, tx *sql.Tx, transactionID TransactionID) (Transaction, error) {
	const q = `
		select transaction
		from ledger_transactions
		where id = $1
		for update;
	`

	var transaction Transaction
	err := tx.QueryRowContext(ctx, q, transactionID).Scan(&transaction)
	if err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func findReversalByOriginalTransactionId(ctx context.Context, tx *sql.Tx, transactionID TransactionID) (TransactionID, error) {
	const q = `
		select original_transaction_id
		from ledger_reversals
		where original_transaction_id = $1;
	`

	var originalTransactionId TransactionID
	err := tx.QueryRowContext(ctx, q, transactionID).Scan(&originalTransactionId)
	if err != nil {
		return 0, nil
	}
	return originalTransactionId, nil
}

func getEntriesByTransactionId(ctx context.Context, tx *sql.Tx, transactionID TransactionID) ([]Entry, error) {
	const q = `
		select * from ledger_entries
		where transaction_id = $1;
	`

	var entries []Entry
	err := tx.QueryRowContext(ctx, q, transactionID).Scan(&entries)
	if err != nil {
		return []Entry{}, nil
	}
	return entries, nil
}

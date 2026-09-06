package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"ledger-db/internal/testdb"
)

func TestVerifyTransactionBalances(t *testing.T) {
	tests := []struct {
		name    string
		entries []int64
		wantErr error
	}{
		{
			name:    "balanced entries",
			entries: []int64{-1000, 1000},
		},
		{
			name:    "unbalanced entries",
			entries: []int64{-1000, 900},
			wantErr: ErrTransactionNotBalanced,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, db := openLedgerstoreIntegrationDB(t)
			resetLedgerstoreIntegrationDB(t, ctx, db)

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()

			transactionID := insertBalanceTestTransaction(t, ctx, tx)
			insertBalanceTestEntries(t, ctx, tx, transactionID, tt.entries)

			err = verifyTransactionBalances(ctx, tx, transactionID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func openLedgerstoreIntegrationDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	return testdb.OpenIntegrationDB(t)
}

func resetLedgerstoreIntegrationDB(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	testdb.ResetIntegrationDB(t, ctx, db)
}

func insertBalanceTestTransaction(t *testing.T, ctx context.Context, tx *sql.Tx) TransactionID {
	t.Helper()

	var fromAccountID int64
	err := tx.QueryRowContext(ctx, `
		insert into ledger_accounts (name, description, currency_code, normal_balance, ledgerable_type, balance)
		values ('Balance Test From', 'Balance test account', 'USD', 'credit', 'internal_account', 0)
		returning id;
	`).Scan(&fromAccountID)
	if err != nil {
		t.Fatal(err)
	}

	var toAccountID int64
	err = tx.QueryRowContext(ctx, `
		insert into ledger_accounts (name, description, currency_code, normal_balance, ledgerable_type, balance)
		values ('Balance Test To', 'Balance test account', 'USD', 'credit', 'internal_account', 0)
		returning id;
	`).Scan(&toAccountID)
	if err != nil {
		t.Fatal(err)
	}

	var transactionID int64
	err = tx.QueryRowContext(ctx, `
		insert into ledger_transactions (
			type,
			idempotency_key,
			from_account_id,
			to_account_id,
			amount,
			currency_code
		)
		values ('transfer', 'balance-test-key', $1, $2, 1000, 'USD')
		returning id;
	`, fromAccountID, toAccountID).Scan(&transactionID)
	if err != nil {
		t.Fatal(err)
	}

	return TransactionID(transactionID)
}

func insertBalanceTestEntries(
	t *testing.T,
	ctx context.Context,
	tx *sql.Tx,
	transactionID TransactionID,
	entries []int64,
) {
	t.Helper()

	var accountID int64
	err := tx.QueryRowContext(ctx, `
		select from_account_id
		from ledger_transactions
		where id = $1;
	`, transactionID).Scan(&accountID)
	if err != nil {
		t.Fatal(err)
	}

	for _, amount := range entries {
		_, err := tx.ExecContext(ctx, `
			insert into ledger_entries (transaction_id, account_id, amount)
			values ($1, $2, $3);
		`, transactionID, accountID, amount)
		if err != nil {
			t.Fatal(err)
		}
	}
}

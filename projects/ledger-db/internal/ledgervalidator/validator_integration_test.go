package ledgervalidator

import (
	"context"
	"database/sql"
	"testing"

	"ledger-db/internal/testdb"
)

func TestValidateHealthyLedger(t *testing.T) {
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)
	seedValidLedger(t, ctx, db)

	result, err := Validate(ctx, db, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if !result.Healthy {
		t.Fatalf("Healthy = false, issues = %#v", result.Issues)
	}
	if result.Summary.AccountCount != 3 {
		t.Fatalf("AccountCount = %d, want 3", result.Summary.AccountCount)
	}
	if result.Summary.TransactionCount != 3 {
		t.Fatalf("TransactionCount = %d, want 3", result.Summary.TransactionCount)
	}
	if result.Summary.EntryCount != 6 {
		t.Fatalf("EntryCount = %d, want 6", result.Summary.EntryCount)
	}
	if result.Summary.ExternalTransferCount != 1 {
		t.Fatalf("ExternalTransferCount = %d, want 1", result.Summary.ExternalTransferCount)
	}
	if result.Summary.ReversalCount != 1 {
		t.Fatalf("ReversalCount = %d, want 1", result.Summary.ReversalCount)
	}
}

func TestValidateReportsAccountBalanceMismatch(t *testing.T) {
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)
	seedValidLedger(t, ctx, db)

	if _, err := db.ExecContext(ctx, `
		update ledger_accounts
		set balance = balance + 1
		where name = 'Alice';
	`); err != nil {
		t.Fatal(err)
	}

	result, err := Validate(ctx, db, Options{})
	if err != nil {
		t.Fatal(err)
	}

	assertIssue(t, result, CheckAccountBalances, "account_balance_mismatch")
}

func TestValidateReportsTransactionBalanceAndShapeIssues(t *testing.T) {
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)
	seedAliceAndBob(t, ctx, db)

	if _, err := db.ExecContext(ctx, `
		insert into ledger_transactions
			(type, idempotency_key, from_account_id, to_account_id, amount, currency_code)
		values
			('transfer', 'bad-transfer', 2, 3, 100, 'USD');

		insert into ledger_entries (transaction_id, account_id, amount)
		values (1, 2, -100);
	`); err != nil {
		t.Fatal(err)
	}

	result, err := Validate(ctx, db, Options{})
	if err != nil {
		t.Fatal(err)
	}

	assertIssue(t, result, CheckTransactionBalances, "transaction_not_balanced")
	assertIssue(t, result, CheckTransactionShape, "transaction_entry_shape_mismatch")
}

func TestValidateReportsExternalTransferMismatch(t *testing.T) {
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)
	seedValidLedger(t, ctx, db)

	if _, err := db.ExecContext(ctx, `
		update external_transfers
		set amount = amount - 1
		where id = 1;
	`); err != nil {
		t.Fatal(err)
	}

	result, err := Validate(ctx, db, Options{})
	if err != nil {
		t.Fatal(err)
	}

	assertIssue(t, result, CheckExternalTransfers, "external_transfer_mismatch")
}

func TestValidateReportsReversalMismatch(t *testing.T) {
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)
	seedAliceAndBob(t, ctx, db)

	if _, err := db.ExecContext(ctx, `
		insert into ledger_transactions
			(type, idempotency_key, from_account_id, to_account_id, amount, currency_code)
		values
			('transfer', 'original-transfer', 2, 3, 100, 'USD'),
			('reversal', 'bad-reversal', 3, 2, 100, 'USD');

		insert into ledger_entries (transaction_id, account_id, amount)
		values
			(1, 2, -100),
			(1, 3, 100),
			(2, 2, 50),
			(2, 3, -50);

		insert into ledger_reversals
			(original_transaction_id, reversal_transaction_id, reason)
		values
			(1, 2, 'bad reversal fixture');
	`); err != nil {
		t.Fatal(err)
	}

	result, err := Validate(ctx, db, Options{})
	if err != nil {
		t.Fatal(err)
	}

	assertIssue(t, result, CheckReversals, "reversal_mismatch")
}

func TestValidateReportsNonPostedTransaction(t *testing.T) {
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)
	seedAliceAndBob(t, ctx, db)

	if _, err := db.ExecContext(ctx, `
		insert into ledger_transactions
			(type, status, idempotency_key, from_account_id, to_account_id, amount, currency_code)
		values
			('transfer', 'pending', 'pending-transfer', 2, 3, 100, 'USD');
	`); err != nil {
		t.Fatal(err)
	}

	result, err := Validate(ctx, db, Options{})
	if err != nil {
		t.Fatal(err)
	}

	assertIssue(t, result, CheckPostedOnlyState, "non_posted_transaction")
}

func assertIssue(t *testing.T, result Result, check CheckName, code string) {
	t.Helper()

	if result.Healthy {
		t.Fatal("Healthy = true, want false")
	}

	for _, issue := range result.Issues {
		if issue.Check == check && issue.Code == code {
			return
		}
	}

	t.Fatalf("missing issue check=%s code=%s in %#v", check, code, result.Issues)
}

func seedValidLedger(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	seedAliceAndBob(t, ctx, db)

	_, err := db.ExecContext(ctx, `
		insert into ledger_transactions
			(type, idempotency_key, from_account_id, to_account_id, amount, currency_code)
		values
			('deposit', 'seed-alice-2000', 1, 2, 2000, 'USD'),
			('transfer', 'alice-sends-bob-1000', 2, 3, 1000, 'USD'),
			('reversal', 'reverse-alice-sends-bob-1000', 3, 2, 1000, 'USD');

		insert into ledger_entries (transaction_id, account_id, amount)
		values
			(1, 1, -2000),
			(1, 2, 2000),
			(2, 2, -1000),
			(2, 3, 1000),
			(3, 2, 1000),
			(3, 3, -1000);

		update ledger_accounts
		set balance = case id
			when 1 then -2000
			when 2 then 2000
			when 3 then 0
			else balance
		end;

		insert into external_transfers (
			direction,
			rail,
			status,
			external_reference,
			user_account_id,
			ledger_transaction_id,
			amount,
			currency_code,
			completed_at
		)
		values (
			'deposit',
			'ach',
			'posted',
			'seed-alice-2000-ext',
			2,
			1,
			2000,
			'USD',
			now()
		);

		insert into ledger_reversals
			(original_transaction_id, reversal_transaction_id, reason)
		values
			(2, 3, 'duplicate transfer');
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func seedAliceAndBob(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		insert into ledger_accounts (name, description, currency_code, balance)
		values
			('Alice', 'Alice Wallet', 'USD', 0),
			('Bob', 'Bob Wallet', 'USD', 0);
	`)
	if err != nil {
		t.Fatal(err)
	}
}

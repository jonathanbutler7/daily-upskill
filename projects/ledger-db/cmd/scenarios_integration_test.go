package ledger

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ledger-db/internal/ledgerstore"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// These tests are Go versions of the original SQL scenarios in db/scenarios.
const defaultTestDSN = "postgresql://ledger_db:password@localhost:5432/ledger_db"

type scenarioAccounts struct {
	cashSettlement int64
	alice          ledgerstore.AccountID
	bob            ledgerstore.AccountID
}

func openIntegrationDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()

	if os.Getenv("LEDGER_DB_INTEGRATION") != "1" {
		t.Skip("set LEDGER_DB_INTEGRATION=1 to run DB-backed scenario tests")
	}

	dsn := os.Getenv("LEDGER_DB_DSN")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	return ctx, db
}

func resetScenarioDB(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	for _, migration := range []string{
		"001_create_ledger_tables.sql",
		"002_create_external_transfers.sql",
		"003_seed_system_accounts.sql",
	} {
		sqlText, err := os.ReadFile(filepath.Join("..", "db", "migrations", migration))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, string(sqlText)); err != nil {
			t.Fatalf("%s: %v", migration, err)
		}
	}
}

func seedAliceAndBob(t *testing.T, ctx context.Context, db *sql.DB) scenarioAccounts {
	t.Helper()

	var accounts scenarioAccounts
	err := db.QueryRowContext(ctx, `
		select id
		from ledger_accounts
		where name = 'Cash Settlement'
			and currency_code = 'USD';
	`).Scan(&accounts.cashSettlement)
	if err != nil {
		t.Fatal(err)
	}

	err = db.QueryRowContext(ctx, `
		insert into ledger_accounts (name, description, currency_code, balance)
		values ('Alice', 'Alice Wallet', 'USD', 0)
		returning id;
	`).Scan(&accounts.alice)
	if err != nil {
		t.Fatal(err)
	}

	err = db.QueryRowContext(ctx, `
		insert into ledger_accounts (name, description, currency_code, balance)
		values ('Bob', 'Bob Wallet', 'USD', 0)
		returning id;
	`).Scan(&accounts.bob)
	if err != nil {
		t.Fatal(err)
	}

	return accounts
}

func assertBalances(t *testing.T, ctx context.Context, db *sql.DB, want map[string]int64) {
	t.Helper()

	rows, err := db.QueryContext(ctx, `
		select name, balance
		from ledger_accounts
		order by id;
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var balance int64
		if err := rows.Scan(&name, &balance); err != nil {
			t.Fatal(err)
		}
		if balance != want[name] {
			t.Fatalf("%s balance = %d, want %d", name, balance, want[name])
		}
		delete(want, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(want) > 0 {
		t.Fatalf("missing balance rows: %v", want)
	}
}

func assertRowCount(t *testing.T, ctx context.Context, db *sql.DB, table string, want int64) {
	t.Helper()

	var got int64
	err := db.QueryRowContext(ctx, "select count(*) from "+table).Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s row count = %d, want %d", table, got, want)
	}
}

func TestScenarioAliceSendsBob(t *testing.T) {
	ctx, db := openIntegrationDB(t)
	resetScenarioDB(t, ctx, db)
	accounts := seedAliceAndBob(t, ctx, db)

	depositID, err := PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		ToAccountID:       accounts.alice,
		TransferAmount:    2000,
		Rail:              "ach",
		ExternalReference: "seed-alice-2000-ext",
		IdempotencyKey:    "seed-alice-2000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if depositID != 1 {
		t.Fatalf("depositID = %d, want 1", depositID)
	}

	transferID, err := PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         1000,
		IdempotencyKey: "alice-sends-bob-1000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if transferID != 2 {
		t.Fatalf("transferID = %d, want 2", transferID)
	}

	assertBalances(t, ctx, db, map[string]int64{
		"Cash Settlement": -2000,
		"Alice":           1000,
		"Bob":             1000,
	})
	assertRowCount(t, ctx, db, "ledger_transactions", 2)
	assertRowCount(t, ctx, db, "ledger_entries", 4)
}

func TestScenarioIdempotency(t *testing.T) {
	ctx, db := openIntegrationDB(t)
	resetScenarioDB(t, ctx, db)
	accounts := seedAliceAndBob(t, ctx, db)

	_, err := PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		ToAccountID:       accounts.alice,
		TransferAmount:    2000,
		Rail:              "ach",
		ExternalReference: "alice-idempotency-seed-ext",
		IdempotencyKey:    "alice-idempotency-seed",
	})
	if err != nil {
		t.Fatal(err)
	}

	firstID, err := PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         1000,
		IdempotencyKey: "same-request",
	})
	if err != nil {
		t.Fatal(err)
	}

	secondID, err := PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         1000,
		IdempotencyKey: "same-request",
	})
	if err != nil {
		t.Fatal(err)
	}
	if secondID != firstID {
		t.Fatalf("secondID = %d, want original transaction %d", secondID, firstID)
	}

	assertBalances(t, ctx, db, map[string]int64{
		"Cash Settlement": -2000,
		"Alice":           1000,
		"Bob":             1000,
	})
	assertRowCount(t, ctx, db, "ledger_transactions", 2)
	assertRowCount(t, ctx, db, "ledger_entries", 4)
}

func TestScenarioInsufficientFunds(t *testing.T) {
	ctx, db := openIntegrationDB(t)
	resetScenarioDB(t, ctx, db)
	accounts := seedAliceAndBob(t, ctx, db)

	_, err := PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		ToAccountID:       accounts.alice,
		TransferAmount:    2000,
		Rail:              "ach",
		ExternalReference: "alice-insufficient-seed-ext",
		IdempotencyKey:    "alice-insufficient-seed",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         3000,
		IdempotencyKey: "same-key",
	})
	if !errors.Is(err, ledgerstore.ErrInsufficientFunds) {
		t.Fatalf("err = %v, want %v", err, ledgerstore.ErrInsufficientFunds)
	}

	assertBalances(t, ctx, db, map[string]int64{
		"Cash Settlement": -2000,
		"Alice":           2000,
		"Bob":             0,
	})
	assertRowCount(t, ctx, db, "ledger_transactions", 1)
	assertRowCount(t, ctx, db, "ledger_entries", 2)
}

func TestScenarioTransferToMissingAccount(t *testing.T) {
	ctx, db := openIntegrationDB(t)
	resetScenarioDB(t, ctx, db)
	accounts := seedAliceAndBob(t, ctx, db)

	_, err := PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		ToAccountID:       accounts.alice,
		TransferAmount:    2000,
		Rail:              "ach",
		ExternalReference: "alice-missing-to-seed-ext",
		IdempotencyKey:    "alice-missing-to-seed",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    999999,
		Amount:         1000,
		IdempotencyKey: "missing-to-account",
	})
	if !errors.Is(err, ledgerstore.ErrToAccountNotFound) {
		t.Fatalf("err = %v, want %v", err, ledgerstore.ErrToAccountNotFound)
	}

	assertBalances(t, ctx, db, map[string]int64{
		"Cash Settlement": -2000,
		"Alice":           2000,
		"Bob":             0,
	})
	assertRowCount(t, ctx, db, "ledger_transactions", 1)
	assertRowCount(t, ctx, db, "ledger_entries", 2)
}

func TestScenarioMismatchedIdempotencyKey(t *testing.T) {
	ctx, db := openIntegrationDB(t)
	resetScenarioDB(t, ctx, db)
	accounts := seedAliceAndBob(t, ctx, db)

	_, err := PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		ToAccountID:       accounts.alice,
		TransferAmount:    5000,
		Rail:              "ach",
		ExternalReference: "alice-mismatch-seed-ext",
		IdempotencyKey:    "alice-mismatch-seed",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         1000,
		IdempotencyKey: "same-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         2000,
		IdempotencyKey: "same-key",
	})
	if !errors.Is(err, ledgerstore.ErrIdempotencyConflict) {
		t.Fatalf("err = %v, want %v", err, ledgerstore.ErrIdempotencyConflict)
	}

	assertBalances(t, ctx, db, map[string]int64{
		"Cash Settlement": -5000,
		"Alice":           4000,
		"Bob":             1000,
	})
	assertRowCount(t, ctx, db, "ledger_transactions", 2)
	assertRowCount(t, ctx, db, "ledger_entries", 4)
}

func TestScenarioStoredAndDerivedBalances(t *testing.T) {
	ctx, db := openIntegrationDB(t)
	resetScenarioDB(t, ctx, db)
	accounts := seedAliceAndBob(t, ctx, db)

	_, err := PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		ToAccountID:       accounts.alice,
		TransferAmount:    2000,
		Rail:              "ach",
		ExternalReference: "alice-balance-check-seed-ext",
		IdempotencyKey:    "alice-balance-check-seed",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         1000,
		IdempotencyKey: "alice-sends-bob-1000",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = PostTransfer(ctx, db, ledgerstore.TransferCommand{
		FromAccountID:  accounts.alice,
		ToAccountID:    accounts.bob,
		Amount:         200,
		IdempotencyKey: "alice-sends-bob-200",
	})
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.QueryContext(ctx, `
		select
			la.name,
			coalesce(sum(le.amount), 0) as derived_balance,
			la.balance as stored_balance
		from ledger_accounts la
		left join ledger_entries le on le.account_id = la.id
		group by la.id, la.name, la.balance
		order by la.id;
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	want := map[string]int64{
		"Cash Settlement": -2000,
		"Alice":           800,
		"Bob":             1200,
	}

	for rows.Next() {
		var name string
		var derivedBalance int64
		var storedBalance int64
		if err := rows.Scan(&name, &derivedBalance, &storedBalance); err != nil {
			t.Fatal(err)
		}
		if derivedBalance != storedBalance {
			t.Fatalf("%s derived balance = %d, stored balance = %d", name, derivedBalance, storedBalance)
		}
		if storedBalance != want[name] {
			t.Fatalf("%s balance = %d, want %d", name, storedBalance, want[name])
		}
		delete(want, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(want) > 0 {
		t.Fatalf("missing balance rows: %v", want)
	}
}

package settlementworker_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"ledger-db/internal/ledgerstore"
	"ledger-db/internal/ledgervalidator"
	"ledger-db/internal/settlementworker"
	"ledger-db/internal/testdb"
)

func fixture(t *testing.T) (context.Context, *sql.DB, int64) {
	t.Helper()
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)
	var user int64
	if err := db.QueryRowContext(ctx, `insert into ledger_accounts
		(name, description, currency_code, normal_balance, ledgerable_type)
		values ('Worker wallet', 'Worker wallet', 'USD', 'credit', 'internal_account') returning id`).Scan(&user); err != nil {
		t.Fatal(err)
	}
	return ctx, db, user
}

func post(t *testing.T, ctx context.Context, db *sql.DB, user int64, key string, amount int64, direction ledgerstore.ExternalTransferDirection, mode ledgerstore.SettlementBalanceMode) ledgerstore.TransactionID {
	t.Helper()
	id, err := ledgerstore.PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		UserAccountID: ledgerstore.AccountID(user), TransferAmount: ledgerstore.Amount(amount), Rail: "ach",
		ExternalReference: ledgerstore.ExternalReference(key), IdempotencyKey: ledgerstore.IdempotencyKey(key),
		ExternalTransferDirection: direction, SettlementBalanceMode: mode,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func balance(t *testing.T, ctx context.Context, db *sql.DB, want int64) {
	t.Helper()
	var got int64
	if err := db.QueryRowContext(ctx, `select balance from ledger_accounts where name = 'Cash Settlement'`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("settlement balance = %d, want %d", got, want)
	}
}

func pending(t *testing.T, ctx context.Context, db *sql.DB, want int64) {
	t.Helper()
	got, err := settlementworker.Pending(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("pending = %d, want %d", got, want)
	}
}

func healthy(t *testing.T, ctx context.Context, db *sql.DB, includePending bool) {
	t.Helper()
	result, err := ledgervalidator.Validate(ctx, db, ledgervalidator.Options{IncludePendingSettlement: includePending})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Healthy {
		t.Fatalf("validation issues: %+v", result.Issues)
	}
}

func TestBatchAppliesOnlyDeferredEntriesAndRollsBackOnFailure(t *testing.T) {
	ctx, db, user := fixture(t)
	post(t, ctx, db, user, "sync", 100, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeSync)
	first := post(t, ctx, db, user, "deferred", 200, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	if replay := post(t, ctx, db, user, "deferred", 200, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync); replay != first {
		t.Fatal("replay created another transaction")
	}
	post(t, ctx, db, user, "withdraw", 50, ledgerstore.ExternalTransferDirectionWithdrawal, ledgerstore.SettlementBalanceModeDeferredAsync)
	balance(t, ctx, db, 100)
	pending(t, ctx, db, 2)
	healthy(t, ctx, db, true)
	exact, err := ledgervalidator.Validate(ctx, db, ledgervalidator.Options{})
	if err != nil || exact.Healthy {
		t.Fatalf("strict validation must detect pending balance: %+v %v", exact, err)
	}

	// Fail after the balance UPDATE, while acknowledging the outbox rows.
	_, err = db.ExecContext(ctx, `create function fail_completion() returns trigger language plpgsql as $$
		begin raise exception 'injected completion failure'; end; $$;
		create trigger fail_completion before update on outbox_events for each row execute function fail_completion();`)
	if err != nil {
		t.Fatal(err)
	}
	w := &settlementworker.Worker{DB: db}
	if _, err := w.ProcessBatch(ctx); err == nil {
		t.Fatal("wanted injected failure")
	}
	balance(t, ctx, db, 100)
	pending(t, ctx, db, 2)
	if _, err := db.ExecContext(ctx, `drop trigger fail_completion on outbox_events; drop function fail_completion();`); err != nil {
		t.Fatal(err)
	}
	if n, err := w.ProcessBatch(ctx); err != nil || n != 2 {
		t.Fatalf("retry = %d, %v", n, err)
	}
	if n, err := w.ProcessBatch(ctx); err != nil || n != 0 {
		t.Fatalf("replay = %d, %v", n, err)
	}
	balance(t, ctx, db, 250)
	pending(t, ctx, db, 0)
	healthy(t, ctx, db, false)
	_, err = db.ExecContext(ctx, `insert into outbox_events (source_type, source_id, event_type)
		select source_type, source_id, event_type from outbox_events where event_type = $1 limit 1`, ledgerstore.SettlementBalanceEvent)
	if err == nil {
		t.Fatal("duplicate entry must be rejected even after processing")
	}
}

func TestDeferredReversalCanBeProcessedBeforeOriginal(t *testing.T) {
	ctx, db, user := fixture(t)
	id := post(t, ctx, db, user, "deposit", 200, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	if _, err := ledgerstore.ReverseTransaction(ctx, db, ledgerstore.ReversalCommand{TransactionID: id, IdempotencyKey: "reverse", Reason: "test"}); err != nil {
		t.Fatal(err)
	}
	pending(t, ctx, db, 2)
	healthy(t, ctx, db, true)
	lock, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.ExecContext(ctx, `select id from outbox_events where event_type = $1 order by id limit 1 for update`, ledgerstore.SettlementBalanceEvent); err != nil {
		t.Fatal(err)
	}
	w := &settlementworker.Worker{DB: db, BatchSize: 1}
	if n, err := w.ProcessBatch(ctx); err != nil || n != 1 {
		t.Fatalf("reverse first = %d, %v", n, err)
	}
	balance(t, ctx, db, -200)
	healthy(t, ctx, db, true)
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	if n, err := w.ProcessBatch(ctx); err != nil || n != 1 {
		t.Fatalf("original = %d, %v", n, err)
	}
	balance(t, ctx, db, 0)
	healthy(t, ctx, db, false)

	// A reversal inherits deferred mode even after the original was processed.
	id = post(t, ctx, db, user, "processed-deposit", 80, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	if _, err := w.ProcessBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := ledgerstore.ReverseTransaction(ctx, db, ledgerstore.ReversalCommand{TransactionID: id, IdempotencyKey: "processed-reverse", Reason: "test"}); err != nil {
		t.Fatal(err)
	}
	balance(t, ctx, db, 80)
	pending(t, ctx, db, 1)
	if _, err := w.ProcessBatch(ctx); err != nil {
		t.Fatal(err)
	}
	healthy(t, ctx, db, false)
}

func TestConcurrentConsumersApplyEveryEntryOnce(t *testing.T) {
	ctx, db, user := fixture(t)
	for i := 0; i < 60; i++ {
		post(t, ctx, db, user, fmt.Sprintf("deposit-%d", i), 7, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := &settlementworker.Worker{DB: db, BatchSize: 3}
			for {
				n, err := w.ProcessBatch(ctx)
				if err != nil {
					t.Error(err)
					return
				}
				if n == 0 {
					return
				}
			}
		}()
	}
	wg.Wait()
	pending(t, ctx, db, 0)
	balance(t, ctx, db, 420)
	healthy(t, ctx, db, false)
}

func TestRunDrainsStartupBacklogAndWakesOnNotifyAfterReconnect(t *testing.T) {
	ctx, db, user := fixture(t)
	post(t, ctx, db, user, "offline", 100, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	dsn := os.Getenv("LEDGER_DB_DSN")
	if dsn == "" {
		dsn = "postgresql://ledger_db:password@localhost:5432/ledger_db"
	}
	w := &settlementworker.Worker{DB: db, DSN: dsn, PollInterval: time.Hour, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- w.Run(runCtx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(3 * time.Second):
			t.Error("worker did not stop")
		}
	})
	waitDrained := func() {
		t.Helper()
		waitCtx, stop := context.WithTimeout(ctx, 3*time.Second)
		defer stop()
		if err := settlementworker.WaitUntilDrained(waitCtx, db); err != nil {
			t.Fatal(err)
		}
	}
	waitDrained()
	post(t, ctx, db, user, "online", 25, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	waitDrained() // Poll interval is an hour: only NOTIFY can wake it in time.
	var killed bool
	if err := db.QueryRowContext(ctx, `select pg_terminate_backend(pid) from pg_stat_activity
		where application_name = 'ledger-settlement-worker' and datname = current_database() limit 1`).Scan(&killed); err != nil || !killed {
		t.Fatalf("terminate listener: %v, %v", killed, err)
	}
	post(t, ctx, db, user, "reconnect", 50, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	waitDrained()
	balance(t, ctx, db, 175)
	healthy(t, ctx, db, false)
}

func TestRunRetriesFailedBatchAndPollsWithoutNotification(t *testing.T) {
	ctx, db, user := fixture(t)
	if _, err := db.ExecContext(ctx, `create or replace function fail_completion() returns trigger language plpgsql as $$
		begin raise exception 'injected completion failure'; end; $$;
		create trigger fail_completion before update on outbox_events for each row execute function fail_completion();
		alter table outbox_events disable trigger settlement_outbox_inserted;`); err != nil {
		t.Fatal(err)
	}
	post(t, ctx, db, user, "retry", 100, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	dsn := os.Getenv("LEDGER_DB_DSN")
	if dsn == "" {
		dsn = "postgresql://ledger_db:password@localhost:5432/ledger_db"
	}
	w := &settlementworker.Worker{DB: db, DSN: dsn, PollInterval: 20 * time.Millisecond, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- w.Run(runCtx) }()
	t.Cleanup(func() { cancel(); <-done })
	deadline := time.After(3 * time.Second)
	for w.Stats().Errors == 0 {
		select {
		case <-deadline:
			t.Fatal("worker never attempted failed batch")
		case <-time.After(5 * time.Millisecond):
		}
	}
	balance(t, ctx, db, 0)
	if _, err := db.ExecContext(ctx, `drop trigger fail_completion on outbox_events; drop function fail_completion();`); err != nil {
		t.Fatal(err)
	}
	waitCtx, stop := context.WithTimeout(ctx, 3*time.Second)
	defer stop()
	if err := settlementworker.WaitUntilDrained(waitCtx, db); err != nil {
		t.Fatal(err)
	}
	post(t, ctx, db, user, "silent", 50, ledgerstore.ExternalTransferDirectionDeposit, ledgerstore.SettlementBalanceModeDeferredAsync)
	if err := settlementworker.WaitUntilDrained(waitCtx, db); err != nil {
		t.Fatal(err)
	}
	balance(t, ctx, db, 150)
	healthy(t, ctx, db, false)
}

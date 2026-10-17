package settlementworker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"ledger-db/internal/ledgerstore"

	"github.com/jackc/pgx/v5"
)

// Worker owns one listener connection. DB should be a small, separate work pool.
// Rows stay pending until the balance update and acknowledgement commit together.
type Worker struct {
	DB           *sql.DB
	DSN          string
	BatchSize    int
	PollInterval time.Duration
	Logger       *slog.Logger

	mu    sync.Mutex
	stats Stats
}

type Stats struct {
	Applied   int64  `json:"applied"`
	Errors    int64  `json:"errors"`
	LastError string `json:"last_error,omitempty"`
}

func (w *Worker) Stats() Stats {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stats
}

func (w *Worker) recordError(err error) {
	w.mu.Lock()
	w.stats.Errors++
	w.stats.LastError = err.Error()
	w.mu.Unlock()
	logger := w.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.Error("settlement worker will retry", "error", err)
}

// ProcessBatch holds row locks until commit. Concurrent consumers skip claimed
// rows; a rollback makes those rows available again without changing balances.
func (w *Worker) ProcessBatch(ctx context.Context) (int, error) {
	batchSize := w.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	tx, err := w.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `
		select o.id, e.account_id
		from outbox_events o
		left join ledger_entries e on e.id = o.source_id
		where o.event_type = $1 and o.processed_at is null
		order by o.id limit $2 for update of o skip locked
	`, ledgerstore.SettlementBalanceEvent, batchSize)
	if err != nil {
		return 0, err
	}
	var ids []int64
	accounts := make(map[int64]bool)
	for rows.Next() {
		var id, accountID int64
		if err := rows.Scan(&id, &accountID); err != nil {
			rows.Close()
			return 0, fmt.Errorf("read settlement entry: %w", err)
		}
		ids = append(ids, id)
		accounts[accountID] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	accountIDs := make([]int64, 0, len(accounts))
	for id := range accounts {
		accountIDs = append(accountIDs, id)
	}
	sort.Slice(accountIDs, func(i, j int) bool { return accountIDs[i] < accountIDs[j] })
	for _, accountID := range accountIDs {
		// This balance is a projection of already-authorized entries. Processing
		// order must not reject a withdrawal because an earlier deposit is pending.
		// PostgreSQL sums bigint as numeric and rejects overflow on assignment.
		_, err := tx.ExecContext(ctx, `
			update ledger_accounts a set balance = a.balance + (
				select sum(case when e.direction = a.normal_balance then e.amount else -e.amount end)
				from outbox_events o join ledger_entries e on e.id = o.source_id
				where o.id = any($1::bigint[]) and e.account_id = a.id
			), lock_version = a.lock_version + 1 where a.id = $2
		`, ids, accountID)
		if err != nil {
			return 0, err
		}
	}
	if _, err := tx.ExecContext(ctx, `update outbox_events set processed_at = now()
		where id = any($1::bigint[])`, ids); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	w.mu.Lock()
	w.stats.Applied += int64(len(ids))
	w.mu.Unlock()
	return len(ids), nil
}

// Run drains on startup, notifications and periodic ticks. Reconnecting always
// establishes LISTEN before scanning the durable backlog. Failed batches remain
// pending and are retried on the next tick.
func (w *Worker) Run(ctx context.Context) error {
	if w.DB == nil || w.DSN == "" {
		return fmt.Errorf("settlement worker requires DB and DSN")
	}
	listenConfig, err := pgx.ParseConfig(w.DSN)
	if err != nil {
		return err
	}
	listenConfig.RuntimeParams["application_name"] = "ledger-settlement-worker"
	interval := w.PollInterval
	if interval <= 0 {
		interval = time.Second
	}
	var conn *pgx.Conn
	closeListener := func() {
		if conn != nil {
			closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			_ = conn.Close(closeCtx)
			cancel()
			conn = nil
		}
	}
	defer closeListener()
	for ctx.Err() == nil {
		if conn == nil {
			connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			var err error
			conn, err = pgx.ConnectConfig(connectCtx, listenConfig)
			if err == nil {
				_, err = conn.Exec(connectCtx, "LISTEN ledger_settlement")
			}
			cancel()
			if err != nil {
				closeListener()
				if ctx.Err() == nil {
					w.recordError(fmt.Errorf("listen: %w", err))
				}
			}
		}
		failed := false
		for ctx.Err() == nil {
			batchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			n, err := w.ProcessBatch(batchCtx)
			cancel()
			if err != nil {
				if ctx.Err() == nil {
					w.recordError(fmt.Errorf("apply batch: %w", err))
				}
				failed = true
				break
			}
			if n == 0 {
				break
			}
		}
		if ctx.Err() != nil {
			break
		}
		if conn == nil || failed {
			timer := time.NewTimer(interval)
			select {
			case <-ctx.Done():
				timer.Stop()
			case <-timer.C:
			}
			continue
		}
		waitCtx, cancel := context.WithTimeout(ctx, interval)
		_, err := conn.WaitForNotification(waitCtx)
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			w.recordError(fmt.Errorf("notification: %w", err))
			closeListener()
		}
	}
	return nil
}

func Pending(ctx context.Context, db *sql.DB) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, `select count(*) from outbox_events
		where event_type = $1 and processed_at is null`, ledgerstore.SettlementBalanceEvent).Scan(&count)
	return count, err
}

// WaitUntilDrained is used after producers stop, while Run continues processing.
func WaitUntilDrained(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		n, err := Pending(ctx, db)
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

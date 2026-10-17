package ledgerstress

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"ledger-db/internal/settlementworker"
)

type SettlementSummary struct {
	Applied         int64  `json:"applied"`
	Pending         int64  `json:"pending"`
	CatchupMS       int64  `json:"catchup_ms"`
	EndToEndMS      int64  `json:"end_to_end_ms"`
	WorkerErrors    int64  `json:"worker_errors"`
	LastWorkerError string `json:"last_worker_error,omitempty"`
	Error           string `json:"error,omitempty"`
}

type runningSettlement struct {
	worker      *settlementworker.Worker
	db          *sql.DB
	cancel      context.CancelFunc
	done        chan error
	once        sync.Once
	seedApplied int64
}

func startSettlement(ctx context.Context, config Config) (*runningSettlement, error) {
	db, err := sql.Open("pgx", config.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	seedPending, err := settlementworker.Pending(ctx, db)
	if err != nil {
		db.Close()
		return nil, err
	}
	workerCtx, cancel := context.WithCancel(ctx)
	w := &settlementworker.Worker{DB: db, DSN: config.DSN,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	s := &runningSettlement{worker: w, db: db, cancel: cancel, done: make(chan error, 1)}
	go func() { s.done <- w.Run(workerCtx) }()
	drainCtx, stop := context.WithTimeout(ctx, config.SettlementDrainTimeout)
	defer stop()
	if err := settlementworker.WaitUntilDrained(drainCtx, db); err != nil {
		s.stop()
		return nil, fmt.Errorf("settlement seed drain: %w (last worker error: %s)", err, w.Stats().LastError)
	}
	s.seedApplied = seedPending
	return s, nil
}

func (s *runningSettlement) stop() {
	s.once.Do(func() { s.cancel(); <-s.done; _ = s.db.Close() })
}

func (s *runningSettlement) finish(ctx context.Context, timeout time.Duration, workloadStarted, requestsFinished time.Time) (*SettlementSummary, error) {
	drainCtx, cancel := context.WithTimeout(ctx, timeout)
	err := settlementworker.WaitUntilDrained(drainCtx, s.db)
	cancel()
	finished := time.Now()
	// Stop before reading final counters, so a just-committed batch is included.
	s.stop()
	stats := s.worker.Stats()
	result := &SettlementSummary{
		Applied: stats.Applied - s.seedApplied, Pending: -1,
		CatchupMS: finished.Sub(requestsFinished).Milliseconds(), EndToEndMS: finished.Sub(workloadStarted).Milliseconds(),
		WorkerErrors: stats.Errors, LastWorkerError: stats.LastError,
	}
	if err != nil {
		result.Error = fmt.Sprintf("settlement drain: %v", err)
	}
	return result, err
}

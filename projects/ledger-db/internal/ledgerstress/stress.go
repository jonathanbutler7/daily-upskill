package ledgerstress

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	ledger "ledger-db/cmd"
	"ledger-db/internal/ledgerschema"
	"ledger-db/internal/ledgerstore"
	"ledger-db/internal/ledgervalidator"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const DefaultDSN = "postgresql://ledger_db:password@localhost:5432/ledger_db"
const defaultMaxOpenConns = 50
const defaultSettlementBuckets = 1

const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

type Config struct {
	DSN                string
	MigrationsDir      string
	Reset              bool
	Accounts           int
	Workers            int
	Operations         int
	SeedBalance        int64
	MaxAmount          int64
	MaxRetries         int
	MaxOpenConns       int
	SettlementBuckets  int
	HotAccounts        int
	HotTransferPercent int
	OperationTimeout   time.Duration
	MinThinkTime       time.Duration
	MaxThinkTime       time.Duration
	ProgressInterval   time.Duration
	ValidationInterval time.Duration
	OutputFormat       string
	Output             io.Writer
}

type Summary struct {
	RunID           string                 `json:"run_id"`
	StartedAt       time.Time              `json:"started_at"`
	FinishedAt      time.Time              `json:"finished_at"`
	DurationMS      int64                  `json:"duration_ms"`
	Config          SummaryConfig          `json:"config"`
	Stats           StatsSnapshot          `json:"stats"`
	DBStats         DBStats                `json:"db_stats"`
	Performance     PerformanceGrade       `json:"performance"`
	FinalValidation ledgervalidator.Result `json:"final_validation"`
	ValidationError string                 `json:"validation_error,omitempty"`
	Healthy         bool                   `json:"healthy"`
}

type SummaryConfig struct {
	Reset              bool  `json:"reset"`
	Accounts           int   `json:"accounts"`
	Workers            int   `json:"workers"`
	Operations         int   `json:"operations"`
	SeedBalance        int64 `json:"seed_balance"`
	MaxAmount          int64 `json:"max_amount"`
	MaxRetries         int   `json:"max_retries"`
	MaxOpenConns       int   `json:"max_open_conns"`
	SettlementBuckets  int   `json:"settlement_buckets"`
	HotAccounts        int   `json:"hot_accounts"`
	HotTransferPercent int   `json:"hot_transfer_percent"`
	OperationTimeoutMS int64 `json:"operation_timeout_ms"`
	MinThinkTimeMS     int64 `json:"min_think_time_ms"`
	MaxThinkTimeMS     int64 `json:"max_think_time_ms"`
}

type StatsSnapshot struct {
	ElapsedMS             int64            `json:"elapsed_ms"`
	Total                 int64            `json:"total"`
	Success               int64            `json:"success"`
	Failure               int64            `json:"failure"`
	RetryAttempts         int64            `json:"retry_attempts"`
	OpsPerSecond          float64          `json:"ops_per_second"`
	ByOperation           map[string]int64 `json:"by_operation"`
	BySuccessfulOperation map[string]int64 `json:"by_successful_operation"`
	MaxLockWaiters        int              `json:"max_lock_waiters"`
	ByErrorCode           map[string]int64 `json:"by_error_code"`
	ByErrorCategory       map[string]int64 `json:"by_error_category"`
	ErrorSamples          []ErrorSample    `json:"error_samples,omitempty"`
	LatencyMS             LatencySnapshot  `json:"latency_ms"`
	LastValidation        *ValidationEvent `json:"last_validation,omitempty"`
}

type ErrorSample struct {
	Code      string `json:"code"`
	Category  string `json:"category"`
	Message   string `json:"message"`
	Operation string `json:"operation"`
	Count     int64  `json:"count"`
}

type LatencySnapshot struct {
	Avg float64 `json:"avg"`
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
	Max float64 `json:"max"`
}

type ValidationEvent struct {
	At         time.Time `json:"at"`
	Healthy    bool      `json:"healthy"`
	IssueCount int       `json:"issue_count"`
	DurationMS int64     `json:"duration_ms"`
	Error      string    `json:"error,omitempty"`
}

type DBStats struct {
	OpenConnections    int   `json:"open_connections"`
	InUse              int   `json:"in_use"`
	Idle               int   `json:"idle"`
	WaitCount          int64 `json:"wait_count"`
	WaitDurationMS     int64 `json:"wait_duration_ms"`
	MaxOpenConnections int   `json:"max_open_connections"`
	LockWaiters        int   `json:"lock_waiters"`
}

type PerformanceGrade struct {
	Grade           string  `json:"grade"`
	Summary         string  `json:"summary"`
	Score           int     `json:"score"`
	ThroughputRPS   float64 `json:"throughput_rps"`
	P95MS           float64 `json:"p95_ms"`
	P99MS           float64 `json:"p99_ms"`
	PoolWaitAvgMS   float64 `json:"pool_wait_avg_ms"`
	PoolWaitPerOpMS float64 `json:"pool_wait_per_op_ms"`
	MaxLockWaiters  int     `json:"max_lock_waiters"`
}

type startEvent struct {
	Type    string        `json:"type"`
	At      time.Time     `json:"at"`
	RunID   string        `json:"run_id"`
	Config  SummaryConfig `json:"config"`
	Message string        `json:"message"`
}

type progressEvent struct {
	Type    string        `json:"type"`
	At      time.Time     `json:"at"`
	RunID   string        `json:"run_id"`
	Target  int           `json:"target"`
	Config  SummaryConfig `json:"config"`
	DBStats DBStats       `json:"db_stats"`
	Stats   StatsSnapshot `json:"stats"`
}

type finalEvent struct {
	Type    string    `json:"type"`
	At      time.Time `json:"at"`
	Summary Summary   `json:"summary"`
}

type workerContext struct {
	runLabel           string
	db                 *sql.DB
	accounts           []ledgerstore.AccountID
	settlementAccounts []ledgerstore.AccountID
	config             Config
	collector          *collector
	sequence           *atomic.Int64
}

type fakeAccountProfile struct {
	Name        string
	Description string
}

func Run(ctx context.Context, config Config) (Summary, error) {
	config = withDefaults(config)
	if err := validateConfig(config); err != nil {
		return Summary{}, err
	}

	runID := uuid.NewString()
	runLabel := runLabelFromID(runID)
	startedAt := time.Now().UTC()
	emit(config, startEvent{
		Type:    "start",
		At:      startedAt,
		RunID:   runID,
		Config:  summaryConfig(config),
		Message: "starting LedgerDB stress run",
	})

	db, err := sql.Open("pgx", config.DSN)
	if err != nil {
		return Summary{}, err
	}
	defer db.Close()
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxOpenConns)

	if err := db.PingContext(ctx); err != nil {
		return Summary{}, err
	}

	if config.Reset {
		if err := ledgerschema.ApplyLocalSchema(ctx, db, config.MigrationsDir); err != nil {
			return Summary{}, fmt.Errorf("reset schema: %w", err)
		}
	}

	accounts, settlementAccounts, err := seedAccounts(ctx, db, runLabel, config)
	if err != nil {
		return Summary{}, err
	}

	collector := newCollector(startedAt)
	done := make(chan struct{})
	var progressWG sync.WaitGroup
	startProgressReporter(ctx, config, runID, db, collector, done, &progressWG)
	startValidationReporter(ctx, config, runID, db, collector, done, &progressWG)

	var sequence atomic.Int64
	workerCtx := workerContext{
		runLabel:           runLabel,
		db:                 db,
		accounts:           accounts,
		settlementAccounts: settlementAccounts,
		config:             config,
		collector:          collector,
		sequence:           &sequence,
	}

	var next atomic.Int64
	var workers sync.WaitGroup
	for workerID := 1; workerID <= config.Workers; workerID++ {
		workers.Add(1)
		go func(workerID int) {
			defer workers.Done()
			runWorker(ctx, workerID, &next, workerCtx)
		}(workerID)
	}
	workers.Wait()
	close(done)
	progressWG.Wait()

	validationStarted := time.Now()
	finalValidation, validationErr := ledgervalidator.Validate(ctx, db, ledgervalidator.Options{})
	validationDuration := time.Since(validationStarted)
	collector.recordValidation(ValidationEvent{
		At:         time.Now().UTC(),
		Healthy:    validationErr == nil && finalValidation.Healthy,
		IssueCount: finalValidation.Summary.IssueCount,
		DurationMS: validationDuration.Milliseconds(),
		Error:      errorString(validationErr),
	})

	finishedAt := time.Now().UTC()
	hasUnexpectedErrors := collector.hasUnexpectedErrors()
	summary := Summary{
		RunID:           runID,
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
		DurationMS:      finishedAt.Sub(startedAt).Milliseconds(),
		Config:          summaryConfig(config),
		Stats:           collector.snapshot(),
		DBStats:         captureDBStats(ctx, db),
		FinalValidation: finalValidation,
		ValidationError: errorString(validationErr),
		Healthy:         validationErr == nil && finalValidation.Healthy && !hasUnexpectedErrors,
	}
	summary.Performance = gradePerformance(summary)

	emit(config, finalEvent{
		Type:    "final",
		At:      finishedAt,
		Summary: summary,
	})

	if validationErr != nil {
		return summary, validationErr
	}
	if !finalValidation.Healthy {
		return summary, fmt.Errorf("final validation found %d issue(s)", finalValidation.Summary.IssueCount)
	}
	if hasUnexpectedErrors {
		return summary, fmt.Errorf("stress run had unexpected errors")
	}

	return summary, nil
}

func withDefaults(config Config) Config {
	if config.DSN == "" {
		config.DSN = DefaultDSN
	}
	if config.MigrationsDir == "" {
		config.MigrationsDir = "db/migrations"
	}
	if config.Accounts == 0 {
		config.Accounts = 25
	}
	if config.Workers == 0 {
		config.Workers = 10
	}
	if config.Operations == 0 {
		config.Operations = 1000
	}
	if config.SeedBalance == 0 {
		config.SeedBalance = 100_000
	}
	if config.MaxAmount == 0 {
		config.MaxAmount = 100
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = min(config.Workers+4, defaultMaxOpenConns)
	}
	if config.SettlementBuckets == 0 {
		config.SettlementBuckets = defaultSettlementBuckets
	}
	if config.OperationTimeout == 0 {
		config.OperationTimeout = 30 * time.Second
	}
	if config.MinThinkTime == 0 {
		config.MinThinkTime = 25 * time.Millisecond
	}
	if config.MaxThinkTime == 0 {
		config.MaxThinkTime = 250 * time.Millisecond
	}
	if config.ProgressInterval == 0 {
		config.ProgressInterval = time.Second
	}
	if config.OutputFormat == "" {
		config.OutputFormat = OutputFormatText
	}
	if config.Output == nil {
		config.Output = io.Discard
	}
	return config
}

func validateConfig(config Config) error {
	if config.Accounts < 2 {
		return fmt.Errorf("accounts must be at least 2")
	}
	if config.Workers < 1 {
		return fmt.Errorf("workers must be at least 1")
	}
	if config.Operations < 1 {
		return fmt.Errorf("operations must be at least 1")
	}
	if config.SeedBalance < 1 {
		return fmt.Errorf("seed balance must be at least 1")
	}
	if config.MaxAmount < 1 {
		return fmt.Errorf("max amount must be at least 1")
	}
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries must not be negative")
	}
	if config.MaxOpenConns < 1 {
		return fmt.Errorf("max open conns must be at least 1")
	}
	if config.SettlementBuckets < 1 {
		return fmt.Errorf("settlement buckets must be at least 1")
	}
	if config.HotAccounts < 0 {
		return fmt.Errorf("hot accounts must not be negative")
	}
	if config.HotAccounts > config.Accounts {
		return fmt.Errorf("hot accounts must be less than or equal to accounts")
	}
	if config.HotTransferPercent < 0 || config.HotTransferPercent > 100 {
		return fmt.Errorf("hot transfer percent must be between 0 and 100")
	}
	if config.HotTransferPercent > 0 && config.HotAccounts < 2 {
		return fmt.Errorf("hot transfer percent requires at least 2 hot accounts")
	}
	if config.MinThinkTime < 0 {
		return fmt.Errorf("min think time must not be negative")
	}
	if config.MaxThinkTime < config.MinThinkTime {
		return fmt.Errorf("max think time must be greater than or equal to min think time")
	}
	if config.OutputFormat != OutputFormatText && config.OutputFormat != OutputFormatJSON {
		return fmt.Errorf("format must be %q or %q", OutputFormatText, OutputFormatJSON)
	}
	return nil
}

func seedAccounts(ctx context.Context, db *sql.DB, runLabel string, config Config) ([]ledgerstore.AccountID, []ledgerstore.AccountID, error) {
	accounts := make([]ledgerstore.AccountID, 0, config.Accounts)
	for i := 1; i <= config.Accounts; i++ {
		profile := fakeStressAccount(i)
		var accountID int64
		err := db.QueryRowContext(ctx, `
			insert into ledger_accounts (name, description, currency_code, balance)
			values ($1, $2, 'USD', 0)
			returning id;
		`,
			profile.Name,
			profile.Description,
		).Scan(&accountID)
		if err != nil {
			return nil, nil, fmt.Errorf("seed account %d: %w", i, err)
		}

		accounts = append(accounts, ledgerstore.AccountID(accountID))
	}

	settlementAccounts, err := seedSettlementBuckets(ctx, db, config)
	if err != nil {
		return nil, nil, err
	}

	for i, accountID := range accounts {
		_, err := ledger.PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
			UserAccountID:             accountID,
			SettlementAccountID:       settlementAccountForIndex(settlementAccounts, i),
			TransferAmount:            ledgerstore.Amount(config.SeedBalance),
			Rail:                      ledgerstore.PaymentRail("ach"),
			ExternalReference:         ledgerstore.ExternalReference(fakeExternalReference(runLabel, "seed", i+1, 0)),
			IdempotencyKey:            ledgerstore.IdempotencyKey(fakeIdempotencyKey(runLabel, "seed", 0, i+1, 0)),
			ExternalTransferDirection: ledgerstore.ExternalTransferDirectionDeposit,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("seed account %d balance: %w", i+1, err)
		}
	}

	return accounts, settlementAccounts, nil
}

func seedSettlementBuckets(ctx context.Context, db *sql.DB, config Config) ([]ledgerstore.AccountID, error) {
	if config.SettlementBuckets <= 1 {
		return nil, nil
	}

	settlementAccounts := make([]ledgerstore.AccountID, 0, config.SettlementBuckets)
	for i := 0; i < config.SettlementBuckets; i++ {
		var accountID int64
		err := db.QueryRowContext(ctx, `
			insert into ledger_accounts (name, description, currency_code, balance)
			values ($1, $2, 'USD', 0)
			returning id;
		`,
			fmt.Sprintf("Cash Settlement ACH %02d", i),
			"Internal stress settlement bucket for ACH write contention tests",
		).Scan(&accountID)
		if err != nil {
			return nil, fmt.Errorf("seed settlement bucket %d: %w", i, err)
		}
		settlementAccounts = append(settlementAccounts, ledgerstore.AccountID(accountID))
	}

	return settlementAccounts, nil
}

func startProgressReporter(ctx context.Context, config Config, runID string, db *sql.DB, collector *collector, done <-chan struct{}, wg *sync.WaitGroup) {
	if config.ProgressInterval < 0 {
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(config.ProgressInterval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case at := <-ticker.C:
				dbStats := captureDBStats(ctx, db)
				collector.recordDBStats(dbStats)
				emit(config, progressEvent{
					Type:    "progress",
					At:      at.UTC(),
					RunID:   runID,
					Target:  config.Operations,
					Config:  summaryConfig(config),
					DBStats: dbStats,
					Stats:   collector.snapshot(),
				})
			}
		}
	}()
}

func startValidationReporter(ctx context.Context, config Config, runID string, db *sql.DB, collector *collector, done <-chan struct{}, wg *sync.WaitGroup) {
	if config.ValidationInterval <= 0 {
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(config.ValidationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case at := <-ticker.C:
				started := time.Now()
				result, err := ledgervalidator.Validate(ctx, db, ledgervalidator.Options{})
				dbStats := captureDBStats(ctx, db)
				collector.recordDBStats(dbStats)
				collector.recordValidation(ValidationEvent{
					At:         at.UTC(),
					Healthy:    err == nil && result.Healthy,
					IssueCount: result.Summary.IssueCount,
					DurationMS: time.Since(started).Milliseconds(),
					Error:      errorString(err),
				})
				emit(config, progressEvent{
					Type:    "validation",
					At:      at.UTC(),
					RunID:   runID,
					Target:  config.Operations,
					Config:  summaryConfig(config),
					DBStats: dbStats,
					Stats:   collector.snapshot(),
				})
			}
		}
	}()
}

func runWorker(ctx context.Context, workerID int, next *atomic.Int64, workerCtx workerContext) {
	random := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

	for {
		operationNumber := int(next.Add(1))
		if operationNumber > workerCtx.config.Operations {
			return
		}
		if ctx.Err() != nil {
			return
		}

		runOperation(ctx, workerID, operationNumber, random, workerCtx)
		sleepThinkTime(ctx, random, workerCtx.config)
	}
}

func sleepThinkTime(ctx context.Context, random *rand.Rand, config Config) {
	if config.MaxThinkTime <= 0 {
		return
	}

	delay := config.MinThinkTime
	if config.MaxThinkTime > config.MinThinkTime {
		delta := config.MaxThinkTime - config.MinThinkTime
		delay += time.Duration(random.Int63n(int64(delta) + 1))
	}
	if delay <= 0 {
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func runOperation(ctx context.Context, workerID int, operationNumber int, random *rand.Rand, workerCtx workerContext) {
	operation := chooseOperation(random)
	amount := ledgerstore.Amount(random.Int63n(workerCtx.config.MaxAmount) + 1)
	sequence := workerCtx.sequence.Add(1)
	idempotencyKey := ledgerstore.IdempotencyKey(fakeIdempotencyKey(workerCtx.runLabel, operation, workerID, operationNumber, sequence))
	externalReference := ledgerstore.ExternalReference(fakeExternalReference(workerCtx.runLabel, operation, operationNumber, sequence))

	started := time.Now()
	err := retry(workerCtx.config.MaxRetries, func() error {
		opCtx, cancel := context.WithTimeout(ctx, workerCtx.config.OperationTimeout)
		defer cancel()

		switch operation {
		case "deposit":
			to := randomAccount(random, workerCtx.accounts)
			_, err := ledger.PostExternalTransfer(opCtx, workerCtx.db, ledgerstore.PostExternalTransferCommand{
				UserAccountID:             to,
				SettlementAccountID:       settlementAccountForIndex(workerCtx.settlementAccounts, int(sequence)),
				TransferAmount:            amount,
				Rail:                      ledgerstore.PaymentRail("ach"),
				ExternalReference:         externalReference,
				IdempotencyKey:            idempotencyKey,
				ExternalTransferDirection: ledgerstore.ExternalTransferDirectionDeposit,
			})
			return err
		case "withdrawal":
			from := randomAccount(random, workerCtx.accounts)
			_, err := ledger.PostExternalTransfer(opCtx, workerCtx.db, ledgerstore.PostExternalTransferCommand{
				UserAccountID:             from,
				SettlementAccountID:       settlementAccountForIndex(workerCtx.settlementAccounts, int(sequence)),
				TransferAmount:            amount,
				Rail:                      ledgerstore.PaymentRail("ach"),
				ExternalReference:         externalReference,
				IdempotencyKey:            idempotencyKey,
				ExternalTransferDirection: ledgerstore.ExternalTransferDirectionWithdrawal,
			})
			return err
		default:
			from, to := randomAccountPair(random, workerCtx.accounts, workerCtx.config)
			_, err := ledger.PostTransfer(opCtx, workerCtx.db, ledgerstore.TransferCommand{
				FromAccountID:  from,
				ToAccountID:    to,
				Amount:         amount,
				IdempotencyKey: idempotencyKey,
			})
			return err
		}
	}, workerCtx.collector)

	workerCtx.collector.record(operation, time.Since(started), err)
}

func retry(maxRetries int, fn func() error, collector *collector) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		info := ledgerstore.ClassifyError(err)
		if !info.Retryable {
			return err
		}
		if attempt < maxRetries {
			collector.recordRetry()
			time.Sleep(time.Duration(attempt+1) * 25 * time.Millisecond)
		}
	}
	return err
}

func chooseOperation(random *rand.Rand) string {
	n := random.Intn(100)
	switch {
	case n < 15:
		return "deposit"
	case n < 30:
		return "withdrawal"
	default:
		return "transfer"
	}
}

func runLabelFromID(runID string) string {
	if len(runID) < 8 {
		return runID
	}
	return runID[:8]
}

func fakeStressAccount(index int) fakeAccountProfile {
	profiles := []fakeAccountProfile{
		{Name: "Austin Market Co", Description: "Merchant wallet for Austin, TX"},
		{Name: "Cedar Park Hardware", Description: "Merchant wallet for Cedar Park, TX"},
		{Name: "Round Rock Fitness", Description: "Merchant wallet for Round Rock, TX"},
		{Name: "Lakeview Books", Description: "Merchant wallet for Lake Travis, TX"},
		{Name: "Oak Hill Studio", Description: "Merchant wallet for Oak Hill, TX"},
		{Name: "Georgetown Outfitters", Description: "Merchant wallet for Georgetown, TX"},
		{Name: "South Congress Coffee", Description: "Merchant wallet for Austin, TX"},
		{Name: "Northline Pharmacy", Description: "Merchant wallet for Austin, TX"},
		{Name: "Brushy Creek Bakery", Description: "Merchant wallet for Cedar Park, TX"},
		{Name: "Westlake Auto Works", Description: "Merchant wallet for West Lake Hills, TX"},
		{Name: "Barton Springs Supply", Description: "Merchant wallet for Austin, TX"},
		{Name: "Pflugerville Pet Supply", Description: "Merchant wallet for Pflugerville, TX"},
	}

	profile := profiles[(index-1)%len(profiles)]
	cycle := (index - 1) / len(profiles)
	if cycle > 0 {
		profile.Name = fmt.Sprintf("%s - Location %d", profile.Name, cycle+1)
	}

	return profile
}

func fakeIdempotencyKey(runLabel string, operation string, workerID int, operationNumber int, sequence int64) string {
	if operation == "seed" {
		return fmt.Sprintf("demo-%s-seed-%04d", runLabel, operationNumber)
	}
	return fmt.Sprintf("demo-%s-%s-w%02d-op%06d-seq%06d", runLabel, operation, workerID, operationNumber, sequence)
}

func fakeExternalReference(runLabel string, operation string, operationNumber int, sequence int64) string {
	if operation == "seed" {
		return fmt.Sprintf("ach-settlement-%s-%04d", runLabel, operationNumber)
	}
	return fmt.Sprintf("ach-%s-%s-%06d-%06d", operation, runLabel, operationNumber, sequence)
}

func settlementAccountForIndex(settlementAccounts []ledgerstore.AccountID, index int) ledgerstore.AccountID {
	if len(settlementAccounts) == 0 {
		return 0
	}
	if index < 0 {
		index = -index
	}
	return settlementAccounts[index%len(settlementAccounts)]
}

func randomAccount(random *rand.Rand, accounts []ledgerstore.AccountID) ledgerstore.AccountID {
	return accounts[random.Intn(len(accounts))]
}

func randomAccountPair(random *rand.Rand, accounts []ledgerstore.AccountID, config Config) (ledgerstore.AccountID, ledgerstore.AccountID) {
	candidates := accounts
	if config.HotTransferPercent > 0 && config.HotAccounts >= 2 && random.Intn(100) < config.HotTransferPercent {
		candidates = accounts[:config.HotAccounts]
	}

	fromIndex := random.Intn(len(candidates))
	toIndex := random.Intn(len(candidates) - 1)
	if toIndex >= fromIndex {
		toIndex++
	}
	return candidates[fromIndex], candidates[toIndex]
}

func summaryConfig(config Config) SummaryConfig {
	return SummaryConfig{
		Reset:              config.Reset,
		Accounts:           config.Accounts,
		Workers:            config.Workers,
		Operations:         config.Operations,
		SeedBalance:        config.SeedBalance,
		MaxAmount:          config.MaxAmount,
		MaxRetries:         config.MaxRetries,
		MaxOpenConns:       config.MaxOpenConns,
		SettlementBuckets:  config.SettlementBuckets,
		HotAccounts:        config.HotAccounts,
		HotTransferPercent: config.HotTransferPercent,
		OperationTimeoutMS: config.OperationTimeout.Milliseconds(),
		MinThinkTimeMS:     config.MinThinkTime.Milliseconds(),
		MaxThinkTimeMS:     config.MaxThinkTime.Milliseconds(),
	}
}

func dbStats(stats sql.DBStats) DBStats {
	return DBStats{
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDurationMS:     stats.WaitDuration.Milliseconds(),
		MaxOpenConnections: stats.MaxOpenConnections,
	}
}

func captureDBStats(ctx context.Context, db *sql.DB) DBStats {
	stats := dbStats(db.Stats())
	lockCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	lockWaiters, err := countLockWaiters(lockCtx, db)
	if err != nil {
		stats.LockWaiters = -1
		return stats
	}
	stats.LockWaiters = lockWaiters
	return stats
}

func countLockWaiters(ctx context.Context, db *sql.DB) (int, error) {
	const q = `
		select count(*)
		from pg_stat_activity
		where datname = current_database()
			and usename = current_user
			and pid <> pg_backend_pid()
			and wait_event_type = 'Lock';
	`

	var count int
	err := db.QueryRowContext(ctx, q).Scan(&count)
	return count, err
}

func gradePerformance(summary Summary) PerformanceGrade {
	stats := summary.Stats
	poolWaitAvg := averageWaitMS(summary.DBStats)
	poolWaitPerOp := waitPerOperationMS(summary.DBStats, stats.Total)
	score := 100

	if !summary.Healthy {
		score -= 40
	}
	if stats.Failure > 0 {
		score -= 15
	}
	if stats.RetryAttempts > 0 {
		score -= 10
	}
	switch {
	case stats.LatencyMS.P99 >= 1000:
		score -= 25
	case stats.LatencyMS.P99 >= 500:
		score -= 20
	case stats.LatencyMS.P99 >= 250:
		score -= 15
	case stats.LatencyMS.P99 >= 100:
		score -= 8
	}
	switch {
	case stats.LatencyMS.P95 >= 250:
		score -= 15
	case stats.LatencyMS.P95 >= 100:
		score -= 10
	case stats.LatencyMS.P95 >= 50:
		score -= 5
	}
	switch {
	case poolWaitPerOp >= 10:
		score -= 15
	case poolWaitPerOp >= 5:
		score -= 10
	case poolWaitPerOp >= 1:
		score -= 5
	}
	if stats.OpsPerSecond < 100 {
		score -= 10
	} else if stats.OpsPerSecond < 500 {
		score -= 5
	}
	if stats.MaxLockWaiters > 0 {
		score -= min(stats.MaxLockWaiters*2, 10)
	}

	if score < 0 {
		score = 0
	}

	grade := letterGrade(score)
	return PerformanceGrade{
		Grade:           grade,
		Summary:         performanceSummary(grade, summary),
		Score:           score,
		ThroughputRPS:   stats.OpsPerSecond,
		P95MS:           stats.LatencyMS.P95,
		P99MS:           stats.LatencyMS.P99,
		PoolWaitAvgMS:   poolWaitAvg,
		PoolWaitPerOpMS: poolWaitPerOp,
		MaxLockWaiters:  stats.MaxLockWaiters,
	}
}

func letterGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func performanceSummary(grade string, summary Summary) string {
	if !summary.Healthy {
		return "correctness failed"
	}
	if summary.Stats.Failure > 0 {
		return "operation failures observed"
	}
	if summary.Stats.LatencyMS.P99 >= 250 {
		return "good throughput; tail latency needs attention"
	}
	if waitPerOperationMS(summary.DBStats, summary.Stats.Total) >= 1 {
		return "good throughput; pool queueing is visible"
	}
	if summary.Stats.MaxLockWaiters > 0 {
		return "correct and fast; lock contention was observed"
	}
	if grade == "A" {
		return "correct, fast, and stable"
	}
	return "correct with some performance pressure"
}

func averageWaitMS(stats DBStats) float64 {
	if stats.WaitCount == 0 {
		return 0
	}
	return float64(stats.WaitDurationMS) / float64(stats.WaitCount)
}

func waitPerOperationMS(stats DBStats, operations int64) float64 {
	if operations == 0 {
		return 0
	}
	return float64(stats.WaitDurationMS) / float64(operations)
}

func emit(config Config, event any) {
	if config.Output == nil {
		return
	}
	if config.OutputFormat == OutputFormatJSON {
		_ = json.NewEncoder(config.Output).Encode(event)
		return
	}
	writeTextEvent(config.Output, event)
}

func writeTextEvent(output io.Writer, event any) {
	switch event := event.(type) {
	case startEvent:
		writeDashboard(output, dashboardData{
			At:     event.At,
			RunID:  event.RunID,
			Config: event.Config,
			Target: event.Config.Operations,
			Mode:   "starting",
		})
	case progressEvent:
		writeDashboard(output, dashboardData{
			At:      event.At,
			RunID:   event.RunID,
			Config:  event.Config,
			Target:  event.Target,
			Stats:   event.Stats,
			DBStats: &event.DBStats,
			Mode:    event.Type,
		})
	case finalEvent:
		writeDashboard(output, dashboardData{
			At:              event.At,
			RunID:           event.Summary.RunID,
			Config:          event.Summary.Config,
			Target:          event.Summary.Config.Operations,
			Stats:           event.Summary.Stats,
			DBStats:         &event.Summary.DBStats,
			Performance:     event.Summary.Performance,
			FinalValidation: &event.Summary.FinalValidation,
			Healthy:         event.Summary.Healthy,
			ValidationError: event.Summary.ValidationError,
			Final:           true,
			Mode:            "final",
		})
	}
}

type dashboardData struct {
	At              time.Time
	RunID           string
	Config          SummaryConfig
	Target          int
	Stats           StatsSnapshot
	DBStats         *DBStats
	Performance     PerformanceGrade
	FinalValidation *ledgervalidator.Result
	Healthy         bool
	ValidationError string
	Final           bool
	Mode            string
}

const (
	ansiReset    = "\033[0m"
	ansiBold     = "\033[1m"
	ansiDim      = "\033[2m"
	ansiPurple   = "\033[35m"
	ansiGreen    = "\033[32m"
	ansiYellow   = "\033[33m"
	ansiRed      = "\033[31m"
	ansiCyan     = "\033[36m"
	ansiClear    = "\033[2J\033[H"
	leftColWidth = 74
)

func writeDashboard(output io.Writer, data dashboardData) {
	stats := data.Stats
	if data.Target == 0 {
		data.Target = data.Config.Operations
	}

	healthText, healthColor := dashboardHealth(data)
	fmt.Fprint(output, ansiClear)
	fmt.Fprintf(output,
		"%s  Health %s%s%s  Run %s  %s  %s\n",
		color(ansiPurple, ansiBold, "LedgerDB Status"),
		healthColor,
		healthText,
		ansiReset,
		shortRunID(data.RunID),
		dim(data.At.Format("15:04:05")),
		dim(fmt.Sprintf("accounts=%d workers=%d ops=%d", data.Config.Accounts, data.Config.Workers, data.Config.Operations)),
	)
	fmt.Fprintf(output, "%s\n", dim("Local Postgres . USD ledger . randomized deposits, withdrawals, and wallet transfers"))
	fmt.Fprintln(output)
	fmt.Fprintln(output, dim("        ______________________________"))
	fmt.Fprintln(output, dim("       | debit       | credit        |"))
	fmt.Fprintln(output, dim("       |_____________|_______________|"))
	fmt.Fprintln(output)

	writeDashboardRow(output, sectionTitle("Workload"), sectionTitle("Latency"))
	writeDashboardRow(output,
		fmt.Sprintf("Ops      %s %6.1f%%  %d/%d", bar(ratio(stats.Total, int64(data.Target)), 28, ansiGreen), ratio(stats.Total, int64(data.Target))*100, stats.Total, data.Target),
		fmt.Sprintf("Avg      %s %8.2f ms", latencyBar(stats.LatencyMS.Avg, data.Config.OperationTimeoutMS), stats.LatencyMS.Avg),
	)
	writeDashboardRow(output,
		fmt.Sprintf("OK       %s %d", bar(ratio(stats.Success, int64(data.Target)), 28, ansiGreen), stats.Success),
		fmt.Sprintf("P50      %s %8.2f ms", latencyBar(stats.LatencyMS.P50, data.Config.OperationTimeoutMS), stats.LatencyMS.P50),
	)
	failColor := ansiGreen
	if stats.Failure > 0 {
		failColor = ansiRed
	}
	writeDashboardRow(output,
		fmt.Sprintf("Failed   %s %d", bar(ratio(stats.Failure, int64(data.Target)), 28, failColor), stats.Failure),
		fmt.Sprintf("P95      %s %8.2f ms", latencyBar(stats.LatencyMS.P95, data.Config.OperationTimeoutMS), stats.LatencyMS.P95),
	)
	writeDashboardRow(output,
		fmt.Sprintf("RPS      %-34s %.1f/sec", smallSpark(stats.OpsPerSecond, float64(max(data.Config.Workers, 1))*8), stats.OpsPerSecond),
		fmt.Sprintf("P99      %s %8.2f ms", latencyBar(stats.LatencyMS.P99, data.Config.OperationTimeoutMS), stats.LatencyMS.P99),
	)
	writeDashboardRow(output,
		fmt.Sprintf("Elapsed  %-34s retries=%d", formatDurationMS(stats.ElapsedMS), stats.RetryAttempts),
		fmt.Sprintf("Max      %s %8.2f ms", latencyBar(stats.LatencyMS.Max, data.Config.OperationTimeoutMS), stats.LatencyMS.Max),
	)
	fmt.Fprintln(output)

	writeDashboardRow(output, sectionTitle("Operations"), sectionTitle("DB Pool"))
	writeDashboardRow(output,
		fmt.Sprintf("Deposit    %s %d", bar(operationRatio(stats, "deposit"), 24, ansiCyan), stats.ByOperation["deposit"]),
		dbPoolLine(data.DBStats, data.Config, "open"),
	)
	writeDashboardRow(output,
		fmt.Sprintf("Transfer   %s %d", bar(operationRatio(stats, "transfer"), 24, ansiCyan), stats.ByOperation["transfer"]),
		dbPoolLine(data.DBStats, data.Config, "in_use"),
	)
	writeDashboardRow(output,
		fmt.Sprintf("Withdrawal %s %d", bar(operationRatio(stats, "withdrawal"), 24, ansiCyan), stats.ByOperation["withdrawal"]),
		dbPoolLine(data.DBStats, data.Config, "waits"),
	)
	fmt.Fprintln(output)

	writeDashboardRow(output, sectionTitle("Validation"), sectionTitle("Errors"))
	writeDashboardRow(output, validationLine(data), errorLine(stats, 0))
	countLines := validationCountLines(data)
	for i, line := range countLines {
		writeDashboardRow(output, line, errorLine(stats, i+1))
	}
	writeDashboardRow(output, validationErrorLine(data), errorLine(stats, len(countLines)+1))

	if data.Final {
		fmt.Fprintln(output)
		fmt.Fprintf(output, "%s\n", color(ansiPurple, ansiBold, "Final"))
		if data.Performance.Grade != "" {
			fmt.Fprintf(output, "  grade:     %s (%s)\n", performanceGradeText(data.Performance), data.Performance.Summary)
			fmt.Fprintf(output, "  signals:   rps=%.1f p95=%.2fms p99=%.2fms pool_wait/op=%.2fms lock_waiters_max=%d\n",
				data.Performance.ThroughputRPS,
				data.Performance.P95MS,
				data.Performance.P99MS,
				data.Performance.PoolWaitPerOpMS,
				data.Performance.MaxLockWaiters,
			)
		}
		fmt.Fprintf(output, "  mix:       %s\n", formatCounts(stats.ByOperation))
		if len(stats.ByErrorCode) > 0 {
			fmt.Fprintf(output, "  errors:    %s\n", formatCounts(stats.ByErrorCode))
		}
	}
}

func dashboardHealth(data dashboardData) (string, string) {
	if data.Final {
		if data.Healthy {
			return "OK", ansiGreen
		}
		return "FAIL", ansiRed
	}
	if data.Stats.Failure > 0 {
		return "CHECK", ansiYellow
	}
	return "RUNNING", ansiGreen
}

func sectionTitle(title string) string {
	return color(ansiPurple, ansiBold, title) + " " + dim(strings.Repeat("-", 48-len(title)))
}

func writeDashboardRow(output io.Writer, left string, right string) {
	fmt.Fprintf(output, "%s  %s\n", padRight(left, leftColWidth), right)
}

func bar(value float64, width int, fillColor string) string {
	value = clamp(value)
	filled := int(value*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	filledText := ""
	if filled > 0 {
		filledText = fillColor + strings.Repeat("#", filled) + ansiReset
	}
	emptyText := ""
	if width > filled {
		emptyText = dim(strings.Repeat("-", width-filled))
	}
	return "[" + filledText + emptyText + "]"
}

func latencyBar(value float64, timeoutMS int64) string {
	limit := float64(timeoutMS)
	if limit <= 0 {
		limit = 1000
	}
	colorCode := ansiGreen
	used := value / limit
	if used >= 0.80 {
		colorCode = ansiRed
	} else if used >= 0.50 {
		colorCode = ansiYellow
	}
	return bar(used, 18, colorCode)
}

func smallSpark(value float64, ceiling float64) string {
	if ceiling <= 0 {
		ceiling = 1
	}
	return bar(value/ceiling, 18, ansiCyan)
}

func operationRatio(stats StatsSnapshot, operation string) float64 {
	return ratio(stats.ByOperation[operation], max64(stats.Total, 1))
}

func dbPoolLine(stats *DBStats, config SummaryConfig, row string) string {
	if stats == nil {
		return dim("waiting for final DB pool snapshot")
	}
	switch row {
	case "open":
		return fmt.Sprintf("Held    %s %d/%d", bar(ratio(int64(stats.OpenConnections), int64(max(config.MaxOpenConns, 1))), 20, ansiGreen), stats.OpenConnections, config.MaxOpenConns)
	case "in_use":
		return fmt.Sprintf("Active  %s %d idle=%d", bar(ratio(int64(stats.InUse), int64(max(config.MaxOpenConns, 1))), 20, ansiYellow), stats.InUse, stats.Idle)
	default:
		lockWaiters := "n/a"
		if stats.LockWaiters >= 0 {
			lockWaiters = fmt.Sprintf("%d", stats.LockWaiters)
		}
		return fmt.Sprintf("Queue   waits=%d avg=%.1fms/op=%.2f lock=%s",
			stats.WaitCount,
			averageWaitMS(*stats),
			waitPerOperationMS(*stats, int64(config.Operations)),
			lockWaiters,
		)
	}
}

func performanceGradeText(performance PerformanceGrade) string {
	colorCode := ansiGreen
	if performance.Grade == "C" || performance.Grade == "D" {
		colorCode = ansiYellow
	}
	if performance.Grade == "F" {
		colorCode = ansiRed
	}
	return color(colorCode, ansiBold, fmt.Sprintf("%s/%d", performance.Grade, performance.Score))
}

func validationLine(data dashboardData) string {
	if data.FinalValidation != nil {
		if data.FinalValidation.Healthy {
			return color(ansiGreen, "", "Final audit healthy")
		}
		return color(ansiRed, "", "Final audit failed")
	}
	if data.Stats.LastValidation != nil {
		event := data.Stats.LastValidation
		if event.Healthy {
			return fmt.Sprintf("%s issues=%d duration=%dms", color(ansiGreen, "", "Last audit healthy"), event.IssueCount, event.DurationMS)
		}
		return fmt.Sprintf("%s issues=%d duration=%dms", color(ansiRed, "", "Last audit failed"), event.IssueCount, event.DurationMS)
	}
	return dim("Final audit runs after workers finish")
}

func validationCountLines(data dashboardData) []string {
	if data.FinalValidation == nil {
		return []string{dim("Use -validation-interval for in-run audits")}
	}

	summary := data.FinalValidation.Summary
	rows := []struct {
		label    string
		expected int64
		actual   int
	}{
		{label: "Accounts", expected: int64(data.Config.Accounts + expectedSettlementAccountCount(data.Config)), actual: summary.AccountCount},
		{label: "Txns", expected: int64(data.Config.Accounts) + data.Stats.Success, actual: summary.TransactionCount},
		{label: "Entries", expected: 2 * (int64(data.Config.Accounts) + data.Stats.Success), actual: summary.EntryCount},
		{label: "External", expected: int64(data.Config.Accounts) + data.Stats.BySuccessfulOperation["deposit"] + data.Stats.BySuccessfulOperation["withdrawal"], actual: summary.ExternalTransferCount},
		{label: "Reversals", expected: 0, actual: summary.ReversalCount},
		{label: "Issues", expected: 0, actual: summary.IssueCount},
	}

	lines := []string{
		validationTableHeader(),
		dim("---------- ---------- ---------- ------"),
	}
	for _, row := range rows {
		expected := fmt.Sprintf("%d", row.expected)
		if !data.Config.Reset {
			expected = "n/a"
		}
		status := validationCountStatus(data.Config.Reset, row.expected, int64(row.actual))
		lines = append(lines, validationTableRow(row.label, expected, row.actual, status))
	}
	return lines
}

func expectedSettlementAccountCount(config SummaryConfig) int {
	if config.SettlementBuckets <= 1 {
		return 1
	}
	return 1 + config.SettlementBuckets
}

func validationTableHeader() string {
	return fmt.Sprintf("%-10s %10s %10s %-6s", "Check", "Expected", "Actual", "Status")
}

func validationTableRow(label string, expected string, actual int, status string) string {
	return fmt.Sprintf("%-10s %10s %10d %-6s", label, expected, actual, status)
}

func validationCountStatus(hasExpected bool, expected int64, actual int64) string {
	if !hasExpected {
		return dim("skip")
	}
	if actual != expected {
		return color(ansiRed, "", "diff")
	}
	if expected == 0 && actual == 0 {
		return color(ansiGreen, "", "clear")
	}
	return color(ansiGreen, "", "ok")
}

func validationErrorLine(data dashboardData) string {
	if data.ValidationError == "" {
		return dim("No validation error")
	}
	return color(ansiRed, "", data.ValidationError)
}

func errorLine(stats StatsSnapshot, index int) string {
	if len(stats.ErrorSamples) == 0 {
		if index == 0 {
			return color(ansiGreen, "", "No operation errors")
		}
		return ""
	}
	if index >= len(stats.ErrorSamples) {
		return ""
	}
	sample := stats.ErrorSamples[index]
	line := fmt.Sprintf("%s %s count=%d op=%s", sample.Category, sample.Code, sample.Count, sample.Operation)
	if sample.Message != "" && (sample.Category == ledgerstore.LedgerErrorCategoryInternal || sample.Code == string(ledgerstore.LedgerErrorCodeInternalUnknown)) {
		line += " msg=" + truncate(sample.Message, 72)
	}
	return line
}

func ratio(value int64, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(value) / float64(total)
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func formatDurationMS(ms int64) string {
	duration := time.Duration(ms) * time.Millisecond
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func shortRunID(runID string) string {
	if len(runID) <= 8 {
		return runID
	}
	return runID[:8]
}

func color(colorCode string, style string, text string) string {
	return style + colorCode + text + ansiReset
}

func dim(text string) string {
	return ansiDim + text + ansiReset
}

func padRight(text string, width int) string {
	visible := visibleLen(text)
	if visible >= width {
		return text
	}
	return text + strings.Repeat(" ", width-visible)
}

func visibleLen(text string) int {
	length := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\033' && i+1 < len(text) && text[i+1] == '[' {
			i += 2
			for i < len(text) && (text[i] < '@' || text[i] > '~') {
				i++
			}
			continue
		}
		length++
	}
	return length
}

func truncate(text string, maxLen int) string {
	if maxLen <= 0 || len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func max64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func formatCounts(counts map[string]int64) string {
	if len(counts) == 0 {
		return "none"
	}

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	text := ""
	for i, key := range keys {
		if i > 0 {
			text += " "
		}
		text += fmt.Sprintf("%s=%d", key, counts[key])
	}
	return text
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type collector struct {
	mu               sync.Mutex
	startedAt        time.Time
	total            int64
	success          int64
	failure          int64
	retryAttempts    int64
	byOperation      map[string]int64
	bySuccess        map[string]int64
	byErrorCode      map[string]int64
	byErrorCategory  map[string]int64
	errorSamples     map[string]*ErrorSample
	latencies        []time.Duration
	lastValidation   *ValidationEvent
	maxLockWaiters   int
	unexpectedErrors int64
}

func newCollector(startedAt time.Time) *collector {
	return &collector{
		startedAt:       startedAt,
		byOperation:     make(map[string]int64),
		bySuccess:       make(map[string]int64),
		byErrorCode:     make(map[string]int64),
		byErrorCategory: make(map[string]int64),
		errorSamples:    make(map[string]*ErrorSample),
	}
}

func (c *collector) record(operation string, latency time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.total++
	c.byOperation[operation]++
	c.latencies = append(c.latencies, latency)
	if err == nil {
		c.success++
		c.bySuccess[operation]++
		return
	}

	c.failure++
	info := ledgerstore.ClassifyError(err)
	code := string(info.Code)
	category := info.Category
	if code == "" {
		code = "unknown"
	}
	if category == "" {
		category = "unknown"
	}
	c.byErrorCode[code]++
	c.byErrorCategory[category]++
	c.recordErrorSample(operation, code, category, info.Message)
	if category != ledgerstore.LedgerErrorCategoryBusiness {
		c.unexpectedErrors++
	}
}

func (c *collector) recordErrorSample(operation string, code string, category string, message string) {
	key := code + "\x00" + operation + "\x00" + message
	if sample, ok := c.errorSamples[key]; ok {
		sample.Count++
		return
	}
	if len(c.errorSamples) >= 10 {
		return
	}
	c.errorSamples[key] = &ErrorSample{
		Code:      code,
		Category:  category,
		Message:   message,
		Operation: operation,
		Count:     1,
	}
}

func (c *collector) recordRetry() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryAttempts++
}

func (c *collector) recordValidation(event ValidationEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastValidation = &event
}

func (c *collector) recordDBStats(stats DBStats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if stats.LockWaiters > c.maxLockWaiters {
		c.maxLockWaiters = stats.LockWaiters
	}
}

func (c *collector) hasUnexpectedErrors() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.unexpectedErrors > 0
}

func (c *collector) snapshot() StatsSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.startedAt)
	opsPerSecond := 0.0
	if elapsed > 0 {
		opsPerSecond = float64(c.total) / elapsed.Seconds()
	}

	return StatsSnapshot{
		ElapsedMS:             elapsed.Milliseconds(),
		Total:                 c.total,
		Success:               c.success,
		Failure:               c.failure,
		RetryAttempts:         c.retryAttempts,
		OpsPerSecond:          opsPerSecond,
		ByOperation:           copyMap(c.byOperation),
		BySuccessfulOperation: copyMap(c.bySuccess),
		MaxLockWaiters:        c.maxLockWaiters,
		ByErrorCode:           copyMap(c.byErrorCode),
		ByErrorCategory:       copyMap(c.byErrorCategory),
		ErrorSamples:          copyErrorSamples(c.errorSamples),
		LatencyMS:             summarizeLatency(c.latencies),
		LastValidation:        copyValidation(c.lastValidation),
	}
}

func copyMap(source map[string]int64) map[string]int64 {
	copied := make(map[string]int64, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}

func copyValidation(source *ValidationEvent) *ValidationEvent {
	if source == nil {
		return nil
	}
	copied := *source
	return &copied
}

func copyErrorSamples(source map[string]*ErrorSample) []ErrorSample {
	samples := make([]ErrorSample, 0, len(source))
	for _, sample := range source {
		samples = append(samples, *sample)
	}
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].Code == samples[j].Code {
			if samples[i].Operation == samples[j].Operation {
				return samples[i].Message < samples[j].Message
			}
			return samples[i].Operation < samples[j].Operation
		}
		return samples[i].Code < samples[j].Code
	})
	return samples
}

func summarizeLatency(latencies []time.Duration) LatencySnapshot {
	if len(latencies) == 0 {
		return LatencySnapshot{}
	}

	values := make([]float64, len(latencies))
	var total float64
	var max float64
	for i, latency := range latencies {
		ms := float64(latency.Microseconds()) / 1000
		values[i] = ms
		total += ms
		if ms > max {
			max = ms
		}
	}
	sort.Float64s(values)

	return LatencySnapshot{
		Avg: total / float64(len(values)),
		P50: percentile(values, 0.50),
		P95: percentile(values, 0.95),
		P99: percentile(values, 0.99),
		Max: max,
	}
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1) * p)
	return values[index]
}

package reconciler

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	goose "github.com/pressly/goose/v3"

	internaldb "weavelab.xyz/payer-sync/internal/db"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	repoRoot := mustRepoRoot()
	_ = godotenv.Load(filepath.Join(repoRoot, ".env"))

	testDatabaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if testDatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "skipping reconciler integration tests: TEST_DATABASE_URL is not set")
		os.Exit(0)
	}
	if err := assertSafeTestDatabaseURL(testDatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "refusing to run reconciler integration tests: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("DATABASE_URL", testDatabaseURL)

	cfg := internaldb.LoadConfigFromEnv()

	sqlDB, err := internaldb.OpenSQLDB(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	if err := goose.Up(sqlDB, filepath.Join(repoRoot, "db", "migrations")); err != nil {
		panic(err)
	}

	testPool, err = internaldb.OpenPool(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer testPool.Close()

	// Truncate all tables touched by the reconciler tests to ensure a clean
	// slate regardless of data left behind by previous interrupted test runs.
	if _, err := testPool.Exec(ctx, `
		TRUNCATE reconciled_payments, state_transitions, job_runs,
		         era_payment_groups, vcc_payment_groups,
		         era_remittances, vcc_files
		CASCADE
	`); err != nil {
		panic(fmt.Sprintf("truncate test tables: %v", err))
	}

	os.Exit(m.Run())
}

func assertSafeTestDatabaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid TEST_DATABASE_URL: %w", err)
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if !strings.Contains(strings.ToLower(dbName), "test") {
		return fmt.Errorf("database name %q must include 'test'", dbName)
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return fmt.Errorf("database host is empty")
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return nil
	}
	if strings.Contains(host, "test") {
		return nil
	}
	return fmt.Errorf("database host %q must be localhost or include 'test'", host)
}

func mustRepoRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filename))
}

type testHarness struct {
	t          *testing.T
	ctx        context.Context
	tx         pgx.Tx
	svc        *Service
	now        time.Time
	locationID string // unique per test; scopes all assertions
}

func newHarness(t *testing.T) *testHarness {
	t.Helper()

	ctx := context.Background()
	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	now := time.Date(2026, 5, 25, 15, 0, 0, 0, time.UTC)
	// Unique locationID per test prevents cross-test data pollution from
	// rows left in the DB by prior interrupted runs.
	locationID := "test-loc-" + sanitiseName(t.Name())
	idGen := newTestIDGen(locationID)

	return &testHarness{
		t:          t,
		ctx:        ctx,
		tx:         tx,
		now:        now,
		locationID: locationID,
		svc: NewService(tx, Config{
			Now:   func() time.Time { return now },
			NewID: idGen,
		}),
	}
}

func sanitiseName(name string) string {
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name) && i < 40; i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			out = append(out, c)
		} else {
			out = append(out, '-')
		}
	}
	return string(out)
}

// newTestIDGen produces IDs scoped to a location so different test transactions
// never produce colliding primary keys visible to each other.
func newTestIDGen(locationID string) func(string) string {
	var n int
	return func(prefix string) string {
		n++
		return fmt.Sprintf("%s-%s-%02d", prefix, locationID[:min(len(locationID), 20)], n)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type eraSeed struct {
	groupID       string
	eraID         string
	locationID    string
	trace         string
	amount        string
	payerName     string
	providerNPI   string
	providerTaxID string
	status        string
	firstReceived time.Time
}

type vccSeed struct {
	groupID         string
	fileID          string
	locationID      string
	trace           string
	paymentID       string
	totalAmount     string
	providerNPI     string
	providerTaxID   string
	cardFingerprint string
	status          string
	firstReceived   time.Time
}

func (h *testHarness) seedERA(s eraSeed) {
	h.t.Helper()

	if s.groupID == "" {
		s.groupID = "era-group-" + s.trace
	}
	if s.eraID == "" {
		s.eraID = "era-" + s.trace
	}
	if s.locationID == "" {
		s.locationID = h.locationID
	}
	if s.payerName == "" {
		s.payerName = "DELTA DENTAL"
	}
	if s.status == "" {
		s.status = "AWAITING_VCC"
	}
	if s.firstReceived.IsZero() {
		s.firstReceived = h.now.Add(-24 * time.Hour)
	}

	_, err := h.tx.Exec(h.ctx, `
		INSERT INTO era_remittances (
			era_id, location_id, payer_name, provider_npi, provider_tax_id,
			bpr_amount, payment_method, trace_number, status, received_at, file_hash, raw_storage_key
		) VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, 'VCC', $7, 'PARSED', $8, $9, $10)
	`, s.eraID, s.locationID, s.payerName, s.providerNPI, s.providerTaxID, s.amount, s.trace, s.firstReceived, "hash-"+s.eraID, "raw/"+s.eraID)
	if err != nil {
		h.t.Fatalf("insert era_remittance: %v", err)
	}

	_, err = h.tx.Exec(h.ctx, `
		INSERT INTO era_payment_groups (
			group_id, era_id, location_id, trace_number, bpr_amount, status, first_received_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, s.groupID, s.eraID, s.locationID, s.trace, s.amount, s.status, s.firstReceived)
	if err != nil {
		h.t.Fatalf("insert era_payment_group: %v", err)
	}
}

func (h *testHarness) seedVCC(s vccSeed) {
	h.t.Helper()

	if s.groupID == "" {
		s.groupID = "vcc-group-" + s.trace
	}
	if s.fileID == "" {
		s.fileID = "vcc-file-" + s.trace
	}
	if s.locationID == "" {
		s.locationID = h.locationID
	}
	if s.paymentID == "" {
		s.paymentID = "payment-" + s.trace
	}
	if s.cardFingerprint == "" {
		s.cardFingerprint = "fp-" + s.trace
	}
	if s.status == "" {
		s.status = "AWAITING_ERA"
	}
	if s.firstReceived.IsZero() {
		s.firstReceived = h.now.Add(-24 * time.Hour)
	}

	_, err := h.tx.Exec(h.ctx, `
		INSERT INTO vcc_files (
			vcc_file_id, location_id, received_at, file_hash, raw_storage_key, row_count, source_filename
		) VALUES ($1, $2, $3, $4, $5, 1, $6)
	`, s.fileID, s.locationID, s.firstReceived, "hash-"+s.fileID, "raw/"+s.fileID, s.fileID+".csv")
	if err != nil {
		h.t.Fatalf("insert vcc_file: %v", err)
	}

	_, err = h.tx.Exec(h.ctx, `
		INSERT INTO vcc_payment_groups (
			group_id, vcc_file_id, location_id, trace_id, payment_id, provider_npi, provider_tax_id,
			card_fingerprint, total_amount, status, is_authoritative, first_received_at
		) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10, TRUE, $11)
	`, s.groupID, s.fileID, s.locationID, s.trace, s.paymentID, s.providerNPI, s.providerTaxID, s.cardFingerprint, s.totalAmount, s.status, s.firstReceived)
	if err != nil {
		h.t.Fatalf("insert vcc_payment_group: %v", err)
	}
}

func (h *testHarness) run() RunResult {
	h.t.Helper()

	result, err := h.svc.Run(h.ctx)
	if err != nil {
		h.t.Fatalf("Run: %v", err)
	}
	return result
}

func (h *testHarness) groupStatus(table, groupID string) string {
	h.t.Helper()

	var status string
	err := h.tx.QueryRow(h.ctx, fmt.Sprintf(`SELECT status FROM %s WHERE group_id = $1`, table), groupID).Scan(&status)
	if err != nil {
		h.t.Fatalf("groupStatus(%s,%s): %v", table, groupID, err)
	}
	return status
}

func (h *testHarness) reconciledPaymentCount() int {
	h.t.Helper()

	var count int
	if err := h.tx.QueryRow(h.ctx,
		`SELECT COUNT(*) FROM reconciled_payments WHERE location_id = $1`, h.locationID,
	).Scan(&count); err != nil {
		h.t.Fatalf("reconciledPaymentCount: %v", err)
	}
	return count
}

func (h *testHarness) latestJobRun() struct {
	runID          string
	jobType        string
	status         string
	recordsMatched int
	finishedAt     sql.NullTime
} {
	h.t.Helper()

	var row struct {
		runID          string
		jobType        string
		status         string
		recordsMatched int
		finishedAt     sql.NullTime
	}
	err := h.tx.QueryRow(h.ctx, `
		SELECT run_id, job_type, status, records_matched, finished_at
		FROM job_runs
		WHERE job_type = 'reconciler'
		ORDER BY started_at DESC
		LIMIT 1
	`).Scan(&row.runID, &row.jobType, &row.status, &row.recordsMatched, &row.finishedAt)
	if err != nil {
		h.t.Fatalf("latestJobRun: %v", err)
	}
	return row
}

func (h *testHarness) jobRunByID(runID string) struct {
	runID          string
	jobType        string
	status         string
	recordsMatched int
	finishedAt     sql.NullTime
} {
	h.t.Helper()

	var row struct {
		runID          string
		jobType        string
		status         string
		recordsMatched int
		finishedAt     sql.NullTime
	}
	err := h.tx.QueryRow(h.ctx, `
		SELECT run_id, job_type, status, records_matched, finished_at
		FROM job_runs
		WHERE run_id = $1
	`, runID).Scan(&row.runID, &row.jobType, &row.status, &row.recordsMatched, &row.finishedAt)
	if err != nil {
		h.t.Fatalf("jobRunByID(%s): %v", runID, err)
	}
	return row
}

func (h *testHarness) transitionCount(entityType, entityID, toState string) int {
	h.t.Helper()

	var count int
	err := h.tx.QueryRow(h.ctx, `
		SELECT COUNT(*)
		FROM state_transitions
		WHERE entity_type = $1 AND entity_id = $2 AND to_state = $3
	`, entityType, entityID, toState).Scan(&count)
	if err != nil {
		h.t.Fatalf("transitionCount: %v", err)
	}
	return count
}

func (h *testHarness) exceptionFields(table, groupID string) (string, sql.NullTime, sql.NullString) {
	h.t.Helper()

	var priorStatus string
	var exceptionAt sql.NullTime
	var reason sql.NullString
	err := h.tx.QueryRow(h.ctx, fmt.Sprintf(`
		SELECT COALESCE(prior_status, ''), exception_at, exception_reason
		FROM %s
		WHERE group_id = $1
	`, table), groupID).Scan(&priorStatus, &exceptionAt, &reason)
	if err != nil {
		h.t.Fatalf("exceptionFields(%s,%s): %v", table, groupID, err)
	}
	return priorStatus, exceptionAt, reason
}

func TestReconciler_LeavesERAAwaitingWhenNoMatchingVCC(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})

	h.run()

	if got := h.groupStatus("era_payment_groups", "era-1"); got != "AWAITING_VCC" {
		t.Fatalf("ERA status = %s, want AWAITING_VCC", got)
	}
	if got := h.reconciledPaymentCount(); got != 0 {
		t.Fatalf("reconciled payments = %d, want 0", got)
	}
}

func TestReconciler_LeavesVCCAwaitingWhenNoMatchingERA(t *testing.T) {
	h := newHarness(t)
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	h.run()

	if got := h.groupStatus("vcc_payment_groups", "vcc-1"); got != "AWAITING_ERA" {
		t.Fatalf("VCC status = %s, want AWAITING_ERA", got)
	}
	if got := h.reconciledPaymentCount(); got != 0 {
		t.Fatalf("reconciled payments = %d, want 0", got)
	}
}

func TestReconciler_MatchesExactTraceAmountAndProvider(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	h.run()

	if got := h.groupStatus("era_payment_groups", "era-1"); got != "MATCHED" {
		t.Fatalf("ERA status = %s, want MATCHED", got)
	}
	if got := h.groupStatus("vcc_payment_groups", "vcc-1"); got != "MATCHED" {
		t.Fatalf("VCC status = %s, want MATCHED", got)
	}
	if got := h.reconciledPaymentCount(); got != 1 {
		t.Fatalf("reconciled payments = %d, want 1", got)
	}
}

func TestReconciler_DoesNotMatchWhenAmountDiffers(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "99.99", providerNPI: "123"})

	h.run()

	if got := h.groupStatus("era_payment_groups", "era-1"); got != "AWAITING_VCC" {
		t.Fatalf("ERA status = %s, want AWAITING_VCC", got)
	}
	if got := h.groupStatus("vcc_payment_groups", "vcc-1"); got != "AWAITING_ERA" {
		t.Fatalf("VCC status = %s, want AWAITING_ERA", got)
	}
	if got := h.reconciledPaymentCount(); got != 0 {
		t.Fatalf("reconciled payments = %d, want 0", got)
	}
}

func TestReconciler_MatchesWhenNPIMatchesAndTaxIDAbsent(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	h.run()

	if got := h.reconciledPaymentCount(); got != 1 {
		t.Fatalf("reconciled payments = %d, want 1", got)
	}
}

func TestReconciler_MatchesWhenTaxIDMatchesAndNPIAbsent(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerTaxID: "11-1111111"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerTaxID: "11-1111111"})

	h.run()

	if got := h.reconciledPaymentCount(); got != 1 {
		t.Fatalf("reconciled payments = %d, want 1", got)
	}
}

func TestReconciler_DoesNotMatchWhenProviderFieldsConflict(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123", providerTaxID: "11-1111111"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123", providerTaxID: "22-2222222"})

	h.run()

	if got := h.reconciledPaymentCount(); got != 0 {
		t.Fatalf("reconciled payments = %d, want 0", got)
	}
}

func TestReconciler_DoesNotMatchWhenBothProviderIdentifiersMissing(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00"})

	h.run()

	if got := h.reconciledPaymentCount(); got != 0 {
		t.Fatalf("reconciled payments = %d, want 0", got)
	}
}

func TestReconciler_DoesNotMatchAcrossLocations(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", locationID: h.locationID + "-A", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-1", locationID: h.locationID + "-B", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	h.run()

	if got := h.reconciledPaymentCount(); got != 0 {
		t.Fatalf("reconciled payments = %d, want 0", got)
	}
}

func TestReconciler_MatchesOutOfOrderWhenERAArrivesFirst(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})

	h.run()
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})
	h.run()

	if got := h.reconciledPaymentCount(); got != 1 {
		t.Fatalf("reconciled payments = %d, want 1", got)
	}
}

func TestReconciler_MatchesOutOfOrderWhenVCCArrivesFirst(t *testing.T) {
	h := newHarness(t)
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	h.run()
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.run()

	if got := h.reconciledPaymentCount(); got != 1 {
		t.Fatalf("reconciled payments = %d, want 1", got)
	}
}

func TestReconciler_ExpiresUnmatchedERAAfterFiveBusinessDays(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{
		groupID:       "era-1",
		trace:         "trace-1",
		amount:        "100.00",
		providerNPI:   "123",
		firstReceived: time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC),
	})

	h.run()

	if got := h.groupStatus("era_payment_groups", "era-1"); got != "EXCEPTION_UNMATCHED" {
		t.Fatalf("ERA status = %s, want EXCEPTION_UNMATCHED", got)
	}
	priorStatus, exceptionAt, reason := h.exceptionFields("era_payment_groups", "era-1")
	if priorStatus != "AWAITING_VCC" {
		t.Fatalf("prior_status = %s, want AWAITING_VCC", priorStatus)
	}
	if !exceptionAt.Valid {
		t.Fatal("exception_at should be set")
	}
	if !reason.Valid || reason.String == "" {
		t.Fatal("exception_reason should be set")
	}
}

func TestReconciler_ExpiresUnmatchedVCCAfterFiveBusinessDays(t *testing.T) {
	h := newHarness(t)
	h.seedVCC(vccSeed{
		groupID:       "vcc-1",
		trace:         "trace-1",
		totalAmount:   "100.00",
		providerNPI:   "123",
		firstReceived: time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC),
	})

	h.run()

	if got := h.groupStatus("vcc_payment_groups", "vcc-1"); got != "EXCEPTION_UNMATCHED" {
		t.Fatalf("VCC status = %s, want EXCEPTION_UNMATCHED", got)
	}
	priorStatus, exceptionAt, reason := h.exceptionFields("vcc_payment_groups", "vcc-1")
	if priorStatus != "AWAITING_ERA" {
		t.Fatalf("prior_status = %s, want AWAITING_ERA", priorStatus)
	}
	if !exceptionAt.Valid {
		t.Fatal("exception_at should be set")
	}
	if !reason.Valid || reason.String == "" {
		t.Fatal("exception_reason should be set")
	}
}

func TestReconciler_DoesNotExpireBeforeFiveBusinessDays(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{
		groupID:       "era-1",
		trace:         "trace-1",
		amount:        "100.00",
		providerNPI:   "123",
		firstReceived: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC),
	})

	h.run()

	if got := h.groupStatus("era_payment_groups", "era-1"); got != "AWAITING_VCC" {
		t.Fatalf("ERA status = %s, want AWAITING_VCC", got)
	}
}

func TestReconciler_IsIdempotentOnRerunAfterMatch(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	h.run()
	h.run()

	if got := h.reconciledPaymentCount(); got != 1 {
		t.Fatalf("reconciled payments = %d, want 1", got)
	}
	if got := h.transitionCount("era_payment_group", "era-1", "MATCHED"); got != 1 {
		t.Fatalf("ERA matched transitions = %d, want 1", got)
	}
	if got := h.transitionCount("vcc_payment_group", "vcc-1", "MATCHED"); got != 1 {
		t.Fatalf("VCC matched transitions = %d, want 1", got)
	}
}

func TestReconciler_IsIdempotentOnRerunAfterExpiration(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{
		groupID:       "era-1",
		trace:         "trace-1",
		amount:        "100.00",
		providerNPI:   "123",
		firstReceived: time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC),
	})

	h.run()
	h.run()

	if got := h.transitionCount("era_payment_group", "era-1", "EXCEPTION_UNMATCHED"); got != 1 {
		t.Fatalf("ERA exception transitions = %d, want 1", got)
	}
}

func TestReconciler_WritesStateTransitionsForMatch(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	h.run()

	if got := h.transitionCount("era_payment_group", "era-1", "MATCHED"); got != 1 {
		t.Fatalf("ERA matched transitions = %d, want 1", got)
	}
	if got := h.transitionCount("vcc_payment_group", "vcc-1", "MATCHED"); got != 1 {
		t.Fatalf("VCC matched transitions = %d, want 1", got)
	}
	var reconciledPaymentID string
	if err := h.tx.QueryRow(h.ctx,
		`SELECT reconciled_payment_id FROM reconciled_payments WHERE location_id = $1 LIMIT 1`,
		h.locationID,
	).Scan(&reconciledPaymentID); err != nil {
		t.Fatalf("lookup reconciled payment: %v", err)
	}
	if got := h.transitionCount("reconciled_payment", reconciledPaymentID, "MATCHED"); got != 1 {
		t.Fatalf("reconciled payment transitions = %d, want 1", got)
	}
}

func TestReconciler_WritesJobRunForReconcilerExecution(t *testing.T) {
	h := newHarness(t)

	result := h.run()
	row := h.jobRunByID(result.RunID)

	if row.runID != result.RunID {
		t.Fatalf("job run id = %s, want %s", row.runID, result.RunID)
	}
	if row.jobType != "reconciler" {
		t.Fatalf("job type = %s, want reconciler", row.jobType)
	}
	if row.status != "success" {
		t.Fatalf("job status = %s, want success", row.status)
	}
	if !row.finishedAt.Valid {
		t.Fatal("finished_at should be set")
	}
}

func TestReconciler_BatchRun_MixedOutcomes(t *testing.T) {
	h := newHarness(t)

	h.seedERA(eraSeed{groupID: "era-match", trace: "trace-match", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-match", trace: "trace-match", totalAmount: "100.00", providerNPI: "123"})

	h.seedERA(eraSeed{groupID: "era-mismatch", trace: "trace-mismatch", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-mismatch", trace: "trace-mismatch", totalAmount: "99.99", providerNPI: "123"})

	stale := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)
	h.seedERA(eraSeed{groupID: "era-stale", trace: "trace-era-stale", amount: "50.00", providerNPI: "123", firstReceived: stale})
	h.seedVCC(vccSeed{groupID: "vcc-stale", trace: "trace-vcc-stale", totalAmount: "75.00", providerNPI: "123", firstReceived: stale})

	result := h.run()

	if got := h.groupStatus("era_payment_groups", "era-match"); got != "MATCHED" {
		t.Fatalf("era-match status = %s, want MATCHED", got)
	}
	if got := h.groupStatus("era_payment_groups", "era-mismatch"); got != "AWAITING_VCC" {
		t.Fatalf("era-mismatch status = %s, want AWAITING_VCC", got)
	}
	if got := h.groupStatus("era_payment_groups", "era-stale"); got != "EXCEPTION_UNMATCHED" {
		t.Fatalf("era-stale status = %s, want EXCEPTION_UNMATCHED", got)
	}
	if got := h.groupStatus("vcc_payment_groups", "vcc-stale"); got != "EXCEPTION_UNMATCHED" {
		t.Fatalf("vcc-stale status = %s, want EXCEPTION_UNMATCHED", got)
	}
	row := h.jobRunByID(result.RunID)
	if row.recordsMatched != result.MatchedCount || row.recordsMatched != 1 {
		t.Fatalf("records_matched = %d, want 1", row.recordsMatched)
	}
}

func TestReconciler_RerunSafety(t *testing.T) {
	h := newHarness(t)
	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	first := h.run()
	second := h.run()

	if first.MatchedCount != 1 {
		t.Fatalf("first matched count = %d, want 1", first.MatchedCount)
	}
	if second.MatchedCount != 0 {
		t.Fatalf("second matched count = %d, want 0", second.MatchedCount)
	}
	if got := h.reconciledPaymentCount(); got != 1 {
		t.Fatalf("reconciled payments = %d, want 1", got)
	}
}

func TestReconciler_EndToEnd_OutOfOrderDelivery(t *testing.T) {
	h := newHarness(t)
	h.seedVCC(vccSeed{groupID: "vcc-1", trace: "trace-1", totalAmount: "100.00", providerNPI: "123"})

	first := h.run()
	if first.MatchedCount != 0 {
		t.Fatalf("first matched count = %d, want 0", first.MatchedCount)
	}

	h.seedERA(eraSeed{groupID: "era-1", trace: "trace-1", amount: "100.00", providerNPI: "123"})
	second := h.run()

	if second.MatchedCount != 1 {
		t.Fatalf("second matched count = %d, want 1", second.MatchedCount)
	}
	if got := h.groupStatus("era_payment_groups", "era-1"); got != "MATCHED" {
		t.Fatalf("ERA status = %s, want MATCHED", got)
	}
	if got := h.groupStatus("vcc_payment_groups", "vcc-1"); got != "MATCHED" {
		t.Fatalf("VCC status = %s, want MATCHED", got)
	}
}

package processor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	goose "github.com/pressly/goose/v3"

	internaldb "weavelab.xyz/payer-sync/internal/db"
)

var testPool *pgxpool.Pool

var testDBAvailable bool

func TestMain(m *testing.M) {
	ctx := context.Background()
	repoRoot := mustRepoRoot()
	_ = godotenv.Load(filepath.Join(repoRoot, ".env"))

	cfg := internaldb.LoadConfigFromEnv()
	sqlDB, err := internaldb.OpenSQLDB(ctx, cfg)
	if err == nil {
		defer sqlDB.Close()
		if goose.SetDialect("postgres") == nil {
			_ = goose.Up(sqlDB, filepath.Join(repoRoot, "db", "migrations"))
		}
		if pool, poolErr := internaldb.OpenPool(ctx, cfg); poolErr == nil {
			testPool = pool
			testDBAvailable = true
			defer pool.Close()
		}
	}

	os.Exit(m.Run())
}

func mustRepoRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filename))
}

func requireIntegration(t *testing.T) {
	t.Helper()
	if !testDBAvailable {
		t.Skip("skipping: no database available (set DATABASE_URL)")
	}
	if os.Getenv("STRIPE_API_KEY") == "" {
		t.Skip("skipping: no STRIPE_API_KEY set")
	}
}

// --- Seed helpers ---

func newTestID(prefix string) string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}

func mustNoErr(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

type seedResult struct {
	paymentID string
}

// createTestPaymentMethod creates a Stripe payment method and returns its ID.
// This simulates what the ingester does at ingestion time.
func createTestPaymentMethod(t *testing.T, _ string) string {
	t.Helper()
	ctx := context.Background()
	proc := newStripeProcessor()
	pm, err := proc.CreatePaymentMethod(ctx, CreatePaymentMethodRequest{
		CardNumber: "tok_visa",
	})
	mustNoErr(t, err, "create test payment method")
	return pm.ID
}

// seedMatchedPayment inserts all rows required for a MATCHED reconciled_payment.
// paymentMethodID is the pre-tokenized Stripe PM ID, stored on vcc_payment_groups.
func seedMatchedPayment(t *testing.T, pool *pgxpool.Pool, paymentMethodID string) seedResult {
	t.Helper()
	ctx := context.Background()

	uid := newTestID("t")
	locationID := "loc-" + uid
	traceNumber := "trace-" + uid
	paymentID := "rp-" + uid
	vccFileID := "vf-" + uid
	vccGroupID := "vg-" + uid
	eraID := "era-" + uid
	eraGroupID := "eg-" + uid
	vccRowID := "vr-" + uid
	now := time.Now().UTC()

	_, err := pool.Exec(ctx, `INSERT INTO vcc_files (vcc_file_id, location_id, received_at, file_hash, raw_storage_key, row_count, source_filename) VALUES ($1,$2,$3,$4,$5,1,$6)`,
		vccFileID, locationID, now, "h-"+vccFileID, "raw/"+vccFileID, vccFileID+".csv")
	mustNoErr(t, err, "insert vcc_file")

	_, err = pool.Exec(ctx, `INSERT INTO vcc_payment_groups (group_id, vcc_file_id, location_id, trace_id, payment_id, card_fingerprint, total_amount, status, is_authoritative, payment_method_id, first_received_at) VALUES ($1,$2,$3,$4,$5,$6,$7,'MATCHED',TRUE,$8,$9)`,
		vccGroupID, vccFileID, locationID, traceNumber, "pay-"+uid, "fp-"+uid, "150.00", paymentMethodID, now)
	mustNoErr(t, err, "insert vcc_payment_group")

	_, err = pool.Exec(ctx, `INSERT INTO vcc_rows (row_id, vcc_file_id, vcc_payment_group_id, location_id, payment_id, trace_id, issue_date, amount, card_fingerprint, last4, expiration_date) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		vccRowID, vccFileID, vccGroupID, locationID, "pay-"+uid, traceNumber, now, "150.00", "fp-"+uid, "4242", "12/2030")
	mustNoErr(t, err, "insert vcc_row")

	_, err = pool.Exec(ctx, `INSERT INTO era_remittances (era_id, location_id, payer_name, bpr_amount, payment_method, trace_number, status, received_at, file_hash, raw_storage_key) VALUES ($1,$2,$3,$4,'VCC',$5,'PARSED',$6,$7,$8)`,
		eraID, locationID, "TEST PAYER", "150.00", traceNumber, now, "h-"+eraID, "raw/"+eraID)
	mustNoErr(t, err, "insert era_remittance")

	_, err = pool.Exec(ctx, `INSERT INTO era_payment_groups (group_id, era_id, location_id, trace_number, bpr_amount, status, first_received_at) VALUES ($1,$2,$3,$4,$5,'MATCHED',$6)`,
		eraGroupID, eraID, locationID, traceNumber, "150.00", now)
	mustNoErr(t, err, "insert era_payment_group")

	_, err = pool.Exec(ctx, `INSERT INTO reconciled_payments (reconciled_payment_id, location_id, era_payment_group_id, vcc_payment_group_id, trace_number, matched_amount, status, matched_at) VALUES ($1,$2,$3,$4,$5,$6,'MATCHED',$7)`,
		paymentID, locationID, eraGroupID, vccGroupID, traceNumber, "150.00", now)
	mustNoErr(t, err, "insert reconciled_payment")

	t.Cleanup(func() {
		bg := context.Background()
		_, _ = pool.Exec(bg, `DELETE FROM processor_attempts WHERE reconciled_payment_id=$1`, paymentID)
		_, _ = pool.Exec(bg, `DELETE FROM state_transitions WHERE entity_id=$1`, paymentID)
		_, _ = pool.Exec(bg, `DELETE FROM reconciled_payments WHERE reconciled_payment_id=$1`, paymentID)
		_, _ = pool.Exec(bg, `DELETE FROM era_payment_groups WHERE group_id=$1`, eraGroupID)
		_, _ = pool.Exec(bg, `DELETE FROM era_remittances WHERE era_id=$1`, eraID)
		_, _ = pool.Exec(bg, `DELETE FROM vcc_rows WHERE row_id=$1`, vccRowID)
		_, _ = pool.Exec(bg, `DELETE FROM vcc_payment_groups WHERE group_id=$1`, vccGroupID)
		_, _ = pool.Exec(bg, `DELETE FROM vcc_files WHERE vcc_file_id=$1`, vccFileID)
	})

	return seedResult{paymentID: paymentID}
}

func newListenConn(t *testing.T) *pgx.Conn {
	t.Helper()
	ctx := context.Background()
	cfg := internaldb.LoadConfigFromEnv()
	dsn, err := cfg.ConnectionString()
	mustNoErr(t, err, "connection string")
	conn, err := pgx.Connect(ctx, dsn)
	mustNoErr(t, err, "connect listen conn")
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func newNotifyConn(t *testing.T) *pgx.Conn {
	t.Helper()
	ctx := context.Background()
	cfg := internaldb.LoadConfigFromEnv()
	dsn, err := cfg.ConnectionString()
	mustNoErr(t, err, "connection string")
	conn, err := pgx.Connect(ctx, dsn)
	mustNoErr(t, err, "connect notify conn")
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func sendNotify(t *testing.T, paymentID string) {
	t.Helper()
	conn := newNotifyConn(t)
	_, err := conn.Exec(context.Background(), "SELECT pg_notify('reconciled_payment_matched', $1)", paymentID)
	mustNoErr(t, err, "send notify")
}

func newTestService(t *testing.T, pool *pgxpool.Pool, proc PaymentProcessor) *Service {
	t.Helper()
	listenConn := newListenConn(t)
	return NewService(pool, Config{
		Processor:  proc,
		ListenConn: listenConn,
		Now:        time.Now,
		NewID:      defaultID,
		MaxRetries: 3,
	})
}

func newStripeProcessor() PaymentProcessor {
	return NewStripeProcessor(StripeConfig{APIKey: os.Getenv("STRIPE_API_KEY")})
}

func startRun(t *testing.T, svc *Service) (cancel context.CancelFunc, runDone <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error, 1)
	go func() { ch <- svc.Run(ctx) }()
	return cancel, ch
}

func pollStatus(t *testing.T, paymentID string, wantStatus PaymentStatus, timeout time.Duration) {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var status string
		err := testPool.QueryRow(ctx, `SELECT status FROM reconciled_payments WHERE reconciled_payment_id=$1`, paymentID).Scan(&status)
		if err == nil && status == wantStatus.String() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	var status string
	_ = testPool.QueryRow(context.Background(), `SELECT status FROM reconciled_payments WHERE reconciled_payment_id=$1`, paymentID).Scan(&status)
	t.Fatalf("timeout waiting for status=%q (got %q) for payment %s", wantStatus, status, paymentID)
}

func countAttempts(t *testing.T, paymentID string) int {
	t.Helper()
	var n int
	err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM processor_attempts WHERE reconciled_payment_id=$1`, paymentID).Scan(&n)
	mustNoErr(t, err, "count attempts")
	return n
}

func getErrorCode(t *testing.T, paymentID string) string {
	t.Helper()
	var code *string
	err := testPool.QueryRow(context.Background(), `SELECT processor_error_code FROM reconciled_payments WHERE reconciled_payment_id=$1`, paymentID).Scan(&code)
	mustNoErr(t, err, "get error code")
	if code == nil {
		return ""
	}
	return *code
}

// --- Integration tests ---

func TestRun_HappyPath(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)
	svc := newTestService(t, testPool, newStripeProcessor())
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)

	pollStatus(t, seed.paymentID, StatusPaymentSucceeded, 10*time.Second)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 1 {
		t.Errorf("want 1 attempt row, got %d", n)
	}

	var transCount int
	err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM state_transitions WHERE entity_id=$1 AND to_state='PAYMENT_SUCCEEDED'`, seed.paymentID).Scan(&transCount)
	mustNoErr(t, err, "count transitions")
	if transCount != 1 {
		t.Errorf("want 1 PAYMENT_SUCCEEDED transition, got %d", transCount)
	}
}

func TestRun_DrainOnStartup(t *testing.T) {
	requireIntegration(t)

	pmSeed1 := createTestPaymentMethod(t, "4242424242424242")
	seed1 := seedMatchedPayment(t, testPool, pmSeed1)
	pmSeed2 := createTestPaymentMethod(t, "4242424242424242")
	seed2 := seedMatchedPayment(t, testPool, pmSeed2)

	svc := newTestService(t, testPool, newStripeProcessor())
	cancel, done := startRun(t, svc)
	defer cancel()

	pollStatus(t, seed1.paymentID, StatusPaymentSucceeded, 15*time.Second)
	pollStatus(t, seed2.paymentID, StatusPaymentSucceeded, 15*time.Second)
	cancel()
	<-done
}

func TestRun_Idempotent_DoesNotDoubleCharge(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)
	svc := newTestService(t, testPool, newStripeProcessor())
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)
	sendNotify(t, seed.paymentID)

	pollStatus(t, seed.paymentID, StatusPaymentSucceeded, 10*time.Second)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 1 {
		t.Errorf("want exactly 1 attempt row (idempotent), got %d", n)
	}
}

func TestRun_SkipsNonMatchedPayments(t *testing.T) {
	requireIntegration(t)

	pmSeed1 := createTestPaymentMethod(t, "4242424242424242")
	seed1 := seedMatchedPayment(t, testPool, pmSeed1)
	pmSeed2 := createTestPaymentMethod(t, "4242424242424242")
	seed2 := seedMatchedPayment(t, testPool, pmSeed2)

	ctx := context.Background()
	_, err := testPool.Exec(ctx, `UPDATE reconciled_payments SET status='PAYMENT_SUCCEEDED' WHERE reconciled_payment_id=$1`, seed1.paymentID)
	mustNoErr(t, err, "set PAYMENT_SUCCEEDED")
	_, err = testPool.Exec(ctx, `UPDATE reconciled_payments SET status='PROCESSING_FAILED' WHERE reconciled_payment_id=$1`, seed2.paymentID)
	mustNoErr(t, err, "set PROCESSING_FAILED")

	svc := newTestService(t, testPool, newStripeProcessor())
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed1.paymentID)
	sendNotify(t, seed2.paymentID)
	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done

	if n := countAttempts(t, seed1.paymentID); n != 0 {
		t.Errorf("want 0 attempts for PAYMENT_SUCCEEDED, got %d", n)
	}
	if n := countAttempts(t, seed2.paymentID); n != 0 {
		t.Errorf("want 0 attempts for PROCESSING_FAILED, got %d", n)
	}
}

func TestRun_TerminalDecline(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)
	// Card decline scenarios are injected via a mock processor since the
	// two-step create→confirm flow must be controlled explicitly in tests.
	proc := &terminalDeclineProcessor{inner: newStripeProcessor(), code: "card_declined"}
	svc := newTestService(t, testPool, proc)
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)

	pollStatus(t, seed.paymentID, StatusProcessingFailed, 10*time.Second)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 1 {
		t.Errorf("want 1 attempt row for terminal decline, got %d", n)
	}
	if code := getErrorCode(t, seed.paymentID); code == "" {
		t.Error("want non-empty processor_error_code for card decline")
	}
}

func TestRun_ExpiredCard(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)
	proc := &terminalDeclineProcessor{inner: newStripeProcessor(), code: "expired_card"}
	svc := newTestService(t, testPool, proc)
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)

	pollStatus(t, seed.paymentID, StatusProcessingFailed, 10*time.Second)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 1 {
		t.Errorf("want 1 attempt row, got %d", n)
	}
}

func TestRun_InvalidCVC(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)
	proc := &terminalDeclineProcessor{inner: newStripeProcessor(), code: "incorrect_cvc"}
	svc := newTestService(t, testPool, proc)
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)

	pollStatus(t, seed.paymentID, StatusProcessingFailed, 10*time.Second)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 1 {
		t.Errorf("want 1 attempt row, got %d", n)
	}
}

// retryOnceProcessor fails ConfirmPaymentIntent once with a retryable error, then delegates.
type retryOnceProcessor struct {
	inner        PaymentProcessor
	confirmCalls int
}

func (r *retryOnceProcessor) CreatePaymentMethod(ctx context.Context, req CreatePaymentMethodRequest) (*PaymentMethod, error) {
	return r.inner.CreatePaymentMethod(ctx, req)
}

func (r *retryOnceProcessor) CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*PaymentIntent, error) {
	return r.inner.CreatePaymentIntent(ctx, req)
}

func (r *retryOnceProcessor) ConfirmPaymentIntent(ctx context.Context, id, key string) (*PaymentIntent, error) {
	r.confirmCalls++
	if r.confirmCalls == 1 {
		return nil, &ProcessorError{Code: "network_failure", Message: "simulated transient"}
	}
	return r.inner.ConfirmPaymentIntent(ctx, id, key)
}

func TestRun_TransientError_RetrySucceeds(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)

	proc := &retryOnceProcessor{inner: newStripeProcessor()}
	svc := newTestService(t, testPool, proc)
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)

	pollStatus(t, seed.paymentID, StatusPaymentSucceeded, 15*time.Second)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 2 {
		t.Errorf("want 2 attempt rows (1 retrying + 1 succeeded), got %d", n)
	}
}

// alwaysFailProcessor always fails ConfirmPaymentIntent with a retryable error.
type alwaysFailProcessor struct {
	inner PaymentProcessor
}

func (r *alwaysFailProcessor) CreatePaymentMethod(ctx context.Context, req CreatePaymentMethodRequest) (*PaymentMethod, error) {
	return r.inner.CreatePaymentMethod(ctx, req)
}

func (r *alwaysFailProcessor) CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*PaymentIntent, error) {
	return r.inner.CreatePaymentIntent(ctx, req)
}

func (r *alwaysFailProcessor) ConfirmPaymentIntent(ctx context.Context, id, key string) (*PaymentIntent, error) {
	return nil, &ProcessorError{Code: "network_failure", Message: "always fails"}
}

// terminalDeclineProcessor fails ConfirmPaymentIntent with a specific terminal (non-retryable) error code.
// Used to test processor behavior when a card is declined, expired, or has wrong CVC.
type terminalDeclineProcessor struct {
	inner PaymentProcessor
	code  string
}

func (p *terminalDeclineProcessor) CreatePaymentMethod(ctx context.Context, req CreatePaymentMethodRequest) (*PaymentMethod, error) {
	return p.inner.CreatePaymentMethod(ctx, req)
}

func (p *terminalDeclineProcessor) CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*PaymentIntent, error) {
	return p.inner.CreatePaymentIntent(ctx, req)
}

func (p *terminalDeclineProcessor) ConfirmPaymentIntent(ctx context.Context, id, key string) (*PaymentIntent, error) {
	return nil, &ProcessorError{Code: p.code, Message: "simulated terminal decline: " + p.code}
}

func TestRun_RetryExhaustion(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)

	proc := &alwaysFailProcessor{inner: newStripeProcessor()}
	svc := newTestService(t, testPool, proc)
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)

	pollStatus(t, seed.paymentID, StatusProcessingFailed, 30*time.Second)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 3 {
		t.Errorf("want 3 attempt rows after retry exhaustion, got %d", n)
	}
}

// --- Pure unit tests (no DB required) ---

func TestProcessorError_IsRetryable(t *testing.T) {
	cases := []struct {
		code      string
		retryable bool
	}{
		{"processor_unavailable", true},
		{"network_failure", true},
		{"rate_limit", true},
		{"card_declined", false},
		{"do_not_honor", false},
		{"expired_card", false},
		{"incorrect_cvc", false},
		{"insufficient_funds", false},
		{"", false},
		{"some_random_code", false},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			e := &ProcessorError{Code: tc.code, Message: "msg"}
			if got := e.IsRetryable(); got != tc.retryable {
				t.Errorf("IsRetryable(%q) = %v, want %v", tc.code, got, tc.retryable)
			}
		})
	}
}

func TestIdempotencyKey_IsDeterministic(t *testing.T) {
	cases := []string{
		"rp-1",
		"rp-2",
		"rp-same-business-data-but-different-row",
		"rp-4",
	}

	seen := map[string]string{}
	for _, paymentID := range cases {
		k1 := computeIdempotencyKey(paymentID)
		k2 := computeIdempotencyKey(paymentID)

		if k1 != k2 {
			t.Errorf("non-deterministic for %s: %q vs %q", paymentID, k1, k2)
		}
		if prev, ok := seen[k1]; ok {
			t.Errorf("collision: key %q used by %q and %q", k1, prev, paymentID)
		}
		seen[k1] = paymentID
	}
}

func TestRun_SkipsPaymentAlreadyInProcessing(t *testing.T) {
	requireIntegration(t)

	pmSeed := createTestPaymentMethod(t, "4242424242424242")
	seed := seedMatchedPayment(t, testPool, pmSeed)

	ctx := context.Background()
	_, err := testPool.Exec(ctx, `UPDATE reconciled_payments SET status='PROCESSING_PAYMENT', idempotency_key='test-key' WHERE reconciled_payment_id=$1`, seed.paymentID)
	mustNoErr(t, err, "set PROCESSING_PAYMENT")

	svc := newTestService(t, testPool, newStripeProcessor())
	cancel, done := startRun(t, svc)
	defer cancel()

	time.Sleep(150 * time.Millisecond)
	sendNotify(t, seed.paymentID)
	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done

	if n := countAttempts(t, seed.paymentID); n != 0 {
		t.Errorf("want 0 attempt rows for already-in-processing payment, got %d", n)
	}
}

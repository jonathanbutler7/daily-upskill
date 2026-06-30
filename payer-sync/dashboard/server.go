package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"weavelab.xyz/payer-sync/internal/appops"
	internaldb "weavelab.xyz/payer-sync/internal/db"
	"weavelab.xyz/payer-sync/processor"
)

const defaultAddr = ":8090"

type Server struct {
	pool       *pgxpool.Pool
	listenConn *pgx.Conn
	log        *slog.Logger
	bus        *eventBus
	state      *dashboardState
	watcher    *progressWatcher
	httpServer *http.Server
}

type dashboardState struct {
	mu         sync.RWMutex
	lastAction actionReport
}

type actionReport struct {
	Kind         string    `json:"kind"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	SeedRoundID  string    `json:"seed_round_id,omitempty"`
	ERACount     int       `json:"era_count,omitempty"`
	VCCCount     int       `json:"vcc_count,omitempty"`
	NewERAFiles  int       `json:"new_era_files,omitempty"`
	NewVCCFiles  int       `json:"new_vcc_files,omitempty"`
	MatchedCount int       `json:"matched_count,omitempty"`
	ExpiredERAs  int       `json:"expired_era_count,omitempty"`
	ExpiredVCCs  int       `json:"expired_vcc_count,omitempty"`
	Error        string    `json:"error,omitempty"`
}

type dashboardSnapshot struct {
	GeneratedAt      time.Time                 `json:"generated_at"`
	Metrics          metricsSnapshot           `json:"metrics"`
	Pipeline         []pipelineStage           `json:"pipeline"`
	LiveFeed         []liveEvent               `json:"live_feed"`
	LastAction       actionReport              `json:"last_action"`
	StateTransitions []stateTransitionSnapshot `json:"state_transitions"`
}

type stateTransitionSnapshot struct {
	TransitionID   string `json:"transition_id"`
	EntityType     string `json:"entity_type"`
	EntityID       string `json:"entity_id"`
	FromState      string `json:"from_state"`
	ToState        string `json:"to_state"`
	TransitionedAt string `json:"transitioned_at"`
	Reason         string `json:"reason"`
}

type metricsSnapshot struct {
	AutoMatchRate         float64 `json:"auto_match_rate"`
	AutoProcessRate       float64 `json:"auto_process_rate"`
	MedianTimeToPostMins  int     `json:"median_time_to_post_mins"`
	ExceptionRatePer1K    float64 `json:"exception_rate_per_1k"`
	TotalProcessedCents   int64   `json:"total_processed_cents"`
	MatchedPayments       int     `json:"matched_payments"`
	ProcessingPayments    int     `json:"processing_payments"`
	OutstandingExceptions int     `json:"outstanding_exceptions"`
	AwaitingMatches       int     `json:"awaiting_matches"`
}

type pipelineStage struct {
	Label       string `json:"label"`
	Count       int    `json:"count"`
	AmountCents int64  `json:"amount_cents"`
	Tone        string `json:"tone"`
}

type liveEvent struct {
	At          string `json:"at"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	TraceNumber string `json:"trace_number"`
	Status      string `json:"status"`
	Tone        string `json:"tone"`
}

type paymentRow struct {
	ReconciledPaymentID      string
	TraceNumber              string
	LocationID               string
	PayerName                string
	Status                   string
	AmountCents              int64
	ProcessorPaymentIntentID string
	UpdatedAt                string
	Progress                 string
}

type eventBus struct {
	mu       sync.Mutex
	seq      int
	buffer   []liveEvent
	watchers map[int]chan liveEvent
}

func newEventBus() *eventBus {
	return &eventBus{watchers: make(map[int]chan liveEvent)}
}

func (b *eventBus) publish(event liveEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.buffer) >= 80 {
		b.buffer = append(b.buffer[1:], event)
	} else {
		b.buffer = append(b.buffer, event)
	}
	for _, ch := range b.watchers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *eventBus) recent() []liveEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	items := make([]liveEvent, len(b.buffer))
	copy(items, b.buffer)
	return items
}

func (b *eventBus) subscribe() (<-chan liveEvent, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seq++
	id := b.seq
	ch := make(chan liveEvent, 16)
	b.watchers[id] = ch
	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.watchers, id)
		close(ch)
	}
}

type progressWatcher struct {
	mu      sync.Mutex
	watched map[string]struct{}
}

func newProgressWatcher() *progressWatcher {
	return &progressWatcher{watched: make(map[string]struct{})}
}

func (w *progressWatcher) claim(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.watched[id]; ok {
		return false
	}
	w.watched[id] = struct{}{}
	return true
}

func (w *progressWatcher) release(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.watched, id)
}

func Run(ctx context.Context) error {
	pool, err := appops.OpenDB(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	dsn, err := internaldb.LoadConfigFromEnv().ConnectionString()
	if err != nil {
		return fmt.Errorf("dashboard dsn: %w", err)
	}
	listenConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("dashboard listen conn: %w", err)
	}
	defer listenConn.Close(ctx)

	return NewServer(pool, listenConn, slog.Default()).Run(ctx)
}

func NewServer(pool *pgxpool.Pool, listenConn *pgx.Conn, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		pool:       pool,
		listenConn: listenConn,
		log:        logger,
		bus:        newEventBus(),
		state:      &dashboardState{},
		watcher:    newProgressWatcher(),
	}
}

func (s *Server) Run(ctx context.Context) error {
	if s.pool == nil {
		return fmt.Errorf("dashboard: db pool is nil")
	}
	if s.listenConn == nil {
		return fmt.Errorf("dashboard: listen connection is nil")
	}
	if _, err := s.listenConn.Exec(ctx, "LISTEN reconciled_payment_matched"); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("dashboard listen: %w", err)
	}

	if err := s.startProcessor(ctx); err != nil {
		return err
	}

	go s.listenLoop(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/snapshot", s.handleSnapshot)
	mux.HandleFunc("/api/ingest-reconcile", s.handleIngestReconcile)
	mux.HandleFunc("/events", s.handleEvents)

	addr := strings.TrimSpace(os.Getenv("DASHBOARD_ADDR"))
	if addr == "" {
		addr = defaultAddr
	}

	s.httpServer = &http.Server{Addr: addr, Handler: mux}
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("dashboard server starting", "addr", addr)
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *Server) startProcessor(ctx context.Context) error {
	stripeAPIKey := strings.TrimSpace(os.Getenv("STRIPE_API_KEY"))
	if stripeAPIKey == "" {
		return fmt.Errorf("dashboard processor: STRIPE_API_KEY env var is required")
	}

	dsn, err := internaldb.LoadConfigFromEnv().ConnectionString()
	if err != nil {
		return fmt.Errorf("dashboard processor dsn: %w", err)
	}
	listenConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("dashboard processor listen conn: %w", err)
	}

	svc := processor.NewService(s.pool, processor.Config{
		Processor:  processor.NewStripeProcessor(processor.StripeConfig{APIKey: stripeAPIKey}),
		ListenConn: listenConn,
	})

	go func() {
		defer listenConn.Close(context.Background())
		err := svc.Run(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			return
		}
		s.log.Error("dashboard embedded processor stopped", "error", err)
		s.publishEvent(liveEvent{
			At:      time.Now().UTC().Format(time.RFC3339),
			Kind:    "processor_error",
			Title:   "Processor stopped",
			Message: err.Error(),
			Tone:    "danger",
		})
	}()

	s.log.Info("dashboard embedded processor started")
	return nil
}

func (s *Server) listenLoop(ctx context.Context) {
	for {
		notification, err := s.listenConn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.log.Warn("dashboard listen stopped", "error", err)
			return
		}
		paymentID := strings.TrimSpace(notification.Payload)
		if paymentID == "" {
			continue
		}
		payment, err := s.loadPayment(ctx, paymentID)
		if err != nil {
			s.log.Warn("dashboard could not load payment", "payment_id", paymentID, "error", err)
			continue
		}
		_ = payment
		if s.watcher.claim(paymentID) {
			go s.trackPayment(ctx, paymentID)
		}
	}
}

func (s *Server) trackPayment(ctx context.Context, paymentID string) {
	defer s.watcher.release(paymentID)
	lastStatus := ""
	ticker := time.NewTicker(750 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()

	for {
		payment, err := s.loadPayment(ctx, paymentID)
		if err != nil {
			return
		}
		if payment.Status != lastStatus {
			lastStatus = payment.Status
		}
		if isTerminalStatus(payment.Status) {
			if payment.Status == "PAYMENT_SUCCEEDED" {
				latest, err := s.loadPayment(ctx, paymentID)
				if err == nil {
					payment = latest
				}
				intentID := strings.TrimSpace(payment.ProcessorPaymentIntentID)
				if intentID == "" {
					intentID = "unknown"
				}
				s.publishEvent(liveEvent{
					At:          time.Now().UTC().Format(time.RFC3339),
					Kind:        "processor_complete",
					Title:       fmt.Sprintf("Processed %s", formatCurrency(payment.AmountCents)),
					Message:     fmt.Sprintf("Payment intent ID: %s", intentID),
					TraceNumber: payment.TraceNumber,
					Status:      payment.Status,
					Tone:        "success",
				})
			}
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-timeout.C:
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) publishEvent(event liveEvent) {
	s.bus.publish(event)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, dashboardHTML)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch, cancel := s.bus.subscribe()
	defer cancel()

	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			payload, _ := json.Marshal(ev)
			_, _ = fmt.Fprintf(w, "event: dashboard\ndata: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.snapshot(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, snapshot, http.StatusOK)
}

func (s *Server) handleIngestReconcile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	started := time.Now().UTC()
	report := actionReport{Kind: "ingest_reconcile", Title: "Ingest and reconcile", StartedAt: started}
	s.state.setAction(report)

	summary, err := appops.Seed(r.Context())
	if err != nil {
		report.FinishedAt = time.Now().UTC()
		report.Error = err.Error()
		report.Message = "seed failed"
		s.state.setAction(report)
		s.publishEvent(liveEvent{At: report.FinishedAt.Format(time.RFC3339), Kind: "action", Title: "Seed failed", Message: err.Error(), Tone: "danger"})
		writeJSON(w, report, http.StatusInternalServerError)
		return
	}
	report.SeedRoundID = summary.SeedRoundID
	report.ERACount = summary.ERACount
	report.VCCCount = summary.VCCCount

	if err := appops.Ingest(r.Context(), s.pool); err != nil {
		report.FinishedAt = time.Now().UTC()
		report.Error = err.Error()
		report.Message = "ingest failed"
		s.state.setAction(report)
		s.publishEvent(liveEvent{At: report.FinishedAt.Format(time.RFC3339), Kind: "action", Title: "Ingest failed", Message: err.Error(), Tone: "danger"})
		writeJSON(w, report, http.StatusInternalServerError)
		return
	}

	result, err := appops.Reconcile(r.Context(), s.pool)
	if err != nil {
		report.FinishedAt = time.Now().UTC()
		report.Error = err.Error()
		report.Message = "reconcile failed"
		s.state.setAction(report)
		s.publishEvent(liveEvent{At: report.FinishedAt.Format(time.RFC3339), Kind: "action", Title: "Reconcile failed", Message: err.Error(), Tone: "danger"})
		writeJSON(w, report, http.StatusInternalServerError)
		return
	}

	newERAs, newVCCs := s.countNewIngests(r.Context(), started)
	matchText := "match"
	if result.MatchedCount != 1 {
		matchText = "matches"
	}
	s.publishEvent(liveEvent{
		At:      time.Now().UTC().Format(time.RFC3339),
		Kind:    "ingest",
		Title:   fmt.Sprintf("Ingested and parsed %d ERA/%d VCC files", newERAs, newVCCs),
		Message: "",
		Tone:    "info",
	})
	s.publishEvent(liveEvent{
		At:      time.Now().UTC().Format(time.RFC3339),
		Kind:    "match",
		Title:   fmt.Sprintf("Found %d %s.", result.MatchedCount, matchText),
		Message: "",
		Tone:    "success",
	})
	report.FinishedAt = time.Now().UTC()
	report.NewERAFiles = newERAs
	report.NewVCCFiles = newVCCs
	report.MatchedCount = result.MatchedCount
	report.ExpiredERAs = result.ExpiredERACount
	report.ExpiredVCCs = result.ExpiredVCCCount
	report.Message = fmt.Sprintf("Ingested and parsed %d ERA/%d VCC files, and found %d %s.", newERAs, newVCCs, result.MatchedCount, matchText)
	s.state.setAction(report)
	writeJSON(w, report, http.StatusOK)
}

func (s *Server) snapshot(ctx context.Context) (dashboardSnapshot, error) {
	metrics, err := s.queryMetrics(ctx)
	if err != nil {
		return dashboardSnapshot{}, err
	}
	pipeline, err := s.queryPipeline(ctx)
	if err != nil {
		return dashboardSnapshot{}, err
	}
	transitions, err := s.queryStateTransitions(ctx)
	if err != nil {
		return dashboardSnapshot{}, err
	}
	return dashboardSnapshot{
		GeneratedAt:      time.Now().UTC(),
		Metrics:          metrics,
		Pipeline:         pipeline,
		LiveFeed:         s.bus.recent(),
		LastAction:       s.state.getAction(),
		StateTransitions: transitions,
	}, nil
}

func (s *Server) queryStateTransitions(ctx context.Context) ([]stateTransitionSnapshot, error) {
	rows, err := s.pool.Query(ctx, `
SELECT transition_id, entity_type, entity_id, COALESCE(from_state, ''), to_state, transitioned_at, COALESCE(reason, '')
FROM state_transitions
ORDER BY transitioned_at DESC
LIMIT 12
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]stateTransitionSnapshot, 0, 12)
	for rows.Next() {
		var item stateTransitionSnapshot
		var transitionedAt time.Time
		if err := rows.Scan(&item.TransitionID, &item.EntityType, &item.EntityID, &item.FromState, &item.ToState, &transitionedAt, &item.Reason); err != nil {
			return nil, err
		}
		item.TransitionedAt = transitionedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Server) queryMetrics(ctx context.Context) (metricsSnapshot, error) {
	var metrics metricsSnapshot
	var matched, processing, exceptions, awaiting, completed int
	var totalProcessed int64
	var medianMins float64
	if err := s.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE status = 'MATCHED'),
  COUNT(*) FILTER (WHERE status = 'PROCESSING_PAYMENT'),
  COUNT(*) FILTER (WHERE status IN ('EXCEPTION', 'PROCESSING_FAILED', 'WRITEBACK_FAILED')),
  COUNT(*) FILTER (WHERE status IN ('AWAITING_VCC', 'AWAITING_ERA')),
  COUNT(*) FILTER (WHERE status IN ('PAYMENT_SUCCEEDED', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED')),
  COALESCE(SUM(CASE WHEN status IN ('POSTED', 'PARTIALLY_POSTED', 'NOTIFIED') THEN (matched_amount * 100)::bigint ELSE 0 END), 0),
  COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (COALESCE(processing_completed_at, updated_at) - matched_at))/60), 0)
FROM reconciled_payments
`).Scan(&matched, &processing, &exceptions, &awaiting, &completed, &totalProcessed, &medianMins); err != nil {
		return metricsSnapshot{}, err
	}
	metrics.MatchedPayments = matched
	metrics.ProcessingPayments = processing
	metrics.OutstandingExceptions = exceptions
	metrics.AwaitingMatches = awaiting
	metrics.TotalProcessedCents = totalProcessed
	metrics.MedianTimeToPostMins = int(medianMins)
	metrics.AutoMatchRate = percentage(matched, matched+awaiting+exceptions)
	metrics.AutoProcessRate = percentage(completed, matched+processing+completed+exceptions)
	metrics.ExceptionRatePer1K = ratePerThousand(exceptions, matched+awaiting+exceptions)
	return metrics, nil
}

func (s *Server) queryPipeline(ctx context.Context) ([]pipelineStage, error) {
	queries := []struct {
		label  string
		count  string
		amount string
		tone   string
	}{
		{"Raw files", `SELECT COALESCE((SELECT COUNT(*) FROM era_remittances), 0) + COALESCE((SELECT COUNT(*) FROM vcc_files), 0)`, `SELECT 0`, "neutral"},
		{"Awaiting match", `SELECT COALESCE((SELECT COUNT(*) FROM era_payment_groups WHERE status = 'AWAITING_VCC'), 0) + COALESCE((SELECT COUNT(*) FROM vcc_payment_groups WHERE status = 'AWAITING_ERA'), 0)`, `SELECT COALESCE((SELECT SUM((bpr_amount * 100)::bigint) FROM era_payment_groups WHERE status = 'AWAITING_VCC'), 0) + COALESCE((SELECT SUM((total_amount * 100)::bigint) FROM vcc_payment_groups WHERE status = 'AWAITING_ERA'), 0)`, "warning"},
		{"Matched", `SELECT COUNT(*) FROM reconciled_payments WHERE status = 'MATCHED'`, `SELECT COALESCE(SUM((matched_amount * 100)::bigint), 0) FROM reconciled_payments WHERE status = 'MATCHED'`, "info"},
		{"Processing", `SELECT COUNT(*) FROM reconciled_payments WHERE status IN ('PROCESSING_PAYMENT', 'PAYMENT_SUCCEEDED')`, `SELECT COALESCE(SUM((matched_amount * 100)::bigint), 0) FROM reconciled_payments WHERE status IN ('PROCESSING_PAYMENT', 'PAYMENT_SUCCEEDED')`, "accent"},
		{"Posted", `SELECT COUNT(*) FROM reconciled_payments WHERE status IN ('WRITING_BACK', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED')`, `SELECT COALESCE(SUM((matched_amount * 100)::bigint), 0) FROM reconciled_payments WHERE status IN ('WRITING_BACK', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED')`, "success"},
	}
	stages := make([]pipelineStage, 0, len(queries))
	for _, item := range queries {
		var count int
		if err := s.pool.QueryRow(ctx, item.count).Scan(&count); err != nil {
			return nil, err
		}
		var amount int64
		if err := s.pool.QueryRow(ctx, item.amount).Scan(&amount); err != nil {
			return nil, err
		}
		stages = append(stages, pipelineStage{Label: item.label, Count: count, AmountCents: amount, Tone: item.tone})
	}
	return stages, nil
}

func (s *Server) loadPayment(ctx context.Context, paymentID string) (paymentRow, error) {
	var item paymentRow
	var amount int64
	var updatedAt time.Time
	err := s.pool.QueryRow(ctx, `
SELECT reconciled_payment_id, trace_number, location_id, COALESCE(payer_name, ''), status, COALESCE((matched_amount * 100)::bigint, 0), COALESCE(processor_payment_intent_id, ''), COALESCE(updated_at, created_at)
FROM reconciled_payments
WHERE reconciled_payment_id = $1
`, paymentID).Scan(&item.ReconciledPaymentID, &item.TraceNumber, &item.LocationID, &item.PayerName, &item.Status, &amount, &item.ProcessorPaymentIntentID, &updatedAt)
	if err != nil {
		return paymentRow{}, err
	}
	item.AmountCents = amount
	item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	item.Progress = progressForStatus(item.Status)
	return item, nil
}

func (s *Server) countNewIngests(ctx context.Context, since time.Time) (int, int) {
	var eraCount, vccCount int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM era_remittances WHERE created_at >= $1`, since).Scan(&eraCount); err != nil {
		return 0, 0
	}
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM vcc_files WHERE created_at >= $1`, since).Scan(&vccCount); err != nil {
		return eraCount, 0
	}
	return eraCount, vccCount
}

func (st *dashboardState) setAction(report actionReport) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.lastAction = report
}

func (st *dashboardState) getAction() actionReport {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.lastAction
}

func writeJSON(w http.ResponseWriter, value any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func percentage(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) * 100 / float64(denominator)
}

func ratePerThousand(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) * 1000 / float64(denominator)
}

func formatCurrency(cents int64) string {
	return fmt.Sprintf("$%.2f", float64(cents)/100)
}

func progressForStatus(status string) string {
	switch status {
	case "MATCHED":
		return "queued"
	case "PROCESSING_PAYMENT":
		return "processing"
	case "PAYMENT_SUCCEEDED":
		return "processor approved"
	case "PROCESSING_FAILED":
		return "processor failed"
	case "POSTED", "PARTIALLY_POSTED", "NOTIFIED":
		return "posted"
	default:
		return strings.ToLower(status)
	}
}

func stageTone(status string) string {
	switch status {
	case "MATCHED", "PAYMENT_SUCCEEDED", "POSTED", "PARTIALLY_POSTED", "NOTIFIED":
		return "success"
	case "PROCESSING_PAYMENT":
		return "info"
	case "PROCESSING_FAILED", "WRITEBACK_FAILED", "EXCEPTION":
		return "danger"
	case "AWAITING_VCC", "AWAITING_ERA":
		return "warning"
	default:
		return "neutral"
	}
}

func isTerminalStatus(status string) bool {
	switch status {
	case "PAYMENT_SUCCEEDED", "PROCESSING_FAILED", "PAYMENT_FAILED", "POSTED", "PARTIALLY_POSTED", "WRITEBACK_FAILED", "NOTIFIED", "EXCEPTION":
		return true
	default:
		return false
	}
}

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>PayerSync Dashboard</title>
  <style>
		html, body { height: 100%; overflow: hidden; }
    :root {
      color-scheme: light;
      --bg: #f4f1ec;
			--card: #ffffff;
			--card-strong: #1d2840;
			--border: #e5e7eb;
      --text: #101827;
      --muted: #64748b;
      --success: #1d9a6c;
      --info: #345bdb;
      --warning: #ca8a04;
      --danger: #dc4f6e;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      color: var(--text);
      background:
        linear-gradient(90deg, #4f77f2 0 14%, transparent 14% 100%),
        linear-gradient(180deg, #f7f4ef, #ede8e1 70%, #ece6de 100%);
			height: 100dvh;
      min-height: 100dvh;
			overflow: hidden;
    }
		.shell { max-width: 1440px; height: 100dvh; max-height: 100dvh; margin: 0 auto; padding: 16px; display:flex; flex-direction:column; gap:12px; overflow:hidden; }
    .topbar, .card {
      background: var(--card); border: 1px solid var(--border); border-radius: 26px;
			box-shadow: 0 12px 30px #d7dce6;
    }
		.topbar { display:flex; justify-content:space-between; gap:20px; align-items:flex-start; padding:16px 18px; margin-bottom:0; flex:0 0 auto; }
		.title h1 { margin:0; font-size:28px; letter-spacing:-0.04em; }
		.title p { margin:6px 0 0; color:var(--muted); max-width:760px; line-height:1.45; font-size:14px; }
    .actions { display:grid; gap:10px; min-width:260px; }
    button { border:0; border-radius:999px; padding:14px 18px; color:white; font-weight:800; cursor:pointer; box-shadow:0 10px 24px rgba(15,23,42,0.14); }
    .primary { background: linear-gradient(135deg, #2f57f7, #133bbd); }
    .success { background: linear-gradient(135deg, #1c9e72, #1d7f5d); }
    .ghost { background: #20293d; }
    .grid { display:grid; gap:16px; }
		.hero { grid-template-columns: minmax(0, 1.65fr) minmax(380px, 1fr); align-items:stretch; flex:1; min-height:0; overflow:hidden; }
		.feed-card { min-height: 0; display:flex; flex-direction:column; }
		.feed-card .panel-body { display:flex; flex-direction:column; min-height:0; }
		.feed-scroll { display:grid; gap:12px; flex:1; min-height:0; overflow:auto; padding-right:4px; }
	    .side-stack { display:grid; gap:12px; min-height:0; height:100%; grid-template-rows: minmax(0, 1fr); overflow:hidden; }
	    .side-stack > .card { min-height:0; display:flex; flex-direction:column; }
	    .side-stack .panel-body { flex:1; min-height:0; display:flex; flex-direction:column; overflow:hidden; }
		.step-row { display:none; }
	.step { padding:14px; border-radius:20px; background:#f5f7fb; border:1px solid #e2e8f0; }
    .step .n { font-size:12px; color:var(--muted); text-transform:uppercase; letter-spacing:.14em; }
    .step .t { margin-top:8px; font-weight:800; }
    .step .d { margin-top:6px; color:var(--muted); font-size:13px; line-height:1.5; }
	.panel-head { padding:18px 22px; border-bottom:1px solid #e5e7eb; display:flex; justify-content:space-between; align-items:center; gap:12px; }
    .panel-head h2 { margin:0; font-size:15px; text-transform:uppercase; letter-spacing:.14em; }
	.panel-body { padding:16px 18px; min-height:0; }
    .badge { display:inline-flex; align-items:center; padding:7px 10px; border-radius:999px; font-size:12px; font-weight:800; }
    .neutral { background:#e2e8f0; color:#334155; }
	.info { background: #dbe7ff; color:#2340a8; }
	.success-badge { background: #d8f3e8; color:#116b4c; }
	.warning { background: #fff2cf; color:#916100; }
	.danger { background: #ffe2e8; color:#b02e57; }
    .card-strong { background: var(--card-strong); color:#f8fafc; }
	.card-strong .panel-head { border-bottom:1px solid #3a4a6d; }
	.card-strong .panel-head h2, .card-strong .small, .card-strong .micro { color: #dce5fb; }
	.feed-item, .metric, .action-line, .pipeline-item { padding:14px 16px; border-radius:18px; background:#fff; border:1px solid #e5e7eb; }
    .feed-item { display:grid; gap:10px; grid-template-columns:110px 1fr auto; align-items:start; }
    .feed-item .meta { color:var(--muted); font-size:12px; font-family:ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
    .feed-item .title { font-weight:800; color:black; }
    .feed-item .body { color:var(--muted); margin-top:6px; line-height:1.45; }
    .feed-item .status { align-self:start; }
	.metrics { grid-template-columns: repeat(4, minmax(0, 1fr)); margin-top:0; }
	.metrics { display:none; }
    .metric .k { color:var(--muted); text-transform:uppercase; letter-spacing:.12em; font-size:12px; }
    .metric .v { margin-top:8px; font-size:32px; font-weight:900; letter-spacing:-.04em; }
    .metric .m { margin-top:8px; color:var(--muted); font-size:13px; line-height:1.45; }
	.pipeline { grid-template-columns: repeat(4, minmax(0, 1fr)); margin-top:12px; }
    .pipeline-item { display:grid; gap:8px; }
    .pipeline-item .k { color:var(--muted); text-transform:uppercase; letter-spacing:.12em; font-size:12px; }
    .pipeline-item .v { font-size:28px; font-weight:900; letter-spacing:-.04em; }
    .pipeline-item .a { color:var(--muted); font-size:13px; }
    .action-lines { display:grid; gap:10px; }
    .action-line { display:flex; justify-content:space-between; gap:16px; align-items:center; }
	#transition-lines { flex:1; min-height:0; overflow:auto; padding-right:4px; }
	.transition-line { border-left: 6px solid #cbd5e1; }
	.transition-gray { border-left-color: #94a3b8; }
	.transition-blue { border-left-color: #3b82f6; }
	.transition-green { border-left-color: #16a34a; }
	.transition-red { border-left-color: #e11d48; }
    .small { font-size:12px; color:var(--muted); }
    .micro { font-size:13px; color:var(--muted); line-height:1.5; }
    .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
    .bold { font-weight:800; }
	@media (max-width: 1180px) { .hero, .metrics, .pipeline { grid-template-columns: 1fr 1fr; } .topbar { flex-direction:column; } .actions { min-width:0; width:100%; grid-template-columns:repeat(2, minmax(0, 1fr)); } .side-stack { height:auto; } }
	@media (max-width: 760px) { html, body { height:auto; overflow:auto; } .hero, .metrics, .pipeline, .feed-item { grid-template-columns:1fr; } .actions { grid-template-columns:1fr; } .feed-scroll { max-height:none; } .shell { height:auto; min-height:100vh; max-height:none; overflow:auto; } body { overflow:auto; } }
  </style>
</head>
<body>
  <div class="shell">
    <div class="topbar">
      <div class="title">
        <h1>PayerSync Dashboard</h1>
        <p>Use the button to poll for a fresh batch, reconcile it, and watch the Processor Feed move in real time.</p>
      </div>
      <div class="actions">
        <button class="success" id="ingest-btn">Get ERA/VCC Data</button>
      </div>
    </div>


	<div class="grid hero">
      <div class="card card-strong feed-card">
        <div class="panel-head">
          <h2>Processor Feed</h2>
          <span class="badge success-badge" id="feed-badge">connecting</span>
        </div>
        <div class="panel-body">
		  <div class="bold" style="font-size:24px; letter-spacing:-0.03em;" id="hero-title">Waiting for the first action.</div>
		  <div class="micro" id="hero-sub" style="margin-top:6px; max-width:760px;">Click <span class="mono">Ingest &amp; Reconcile</span> to run seed, ingest, and reconcile, then watch live processor progress.</div>
		  <div class="feed-scroll" id="feed" style="margin-top:12px;"></div>
        </div>
      </div>

      <div class="side-stack">
		<div class="card">
          <div class="panel-head">
			<h2>State Transitions</h2>
			<span class="badge neutral">audit trail</span>
          </div>
          <div class="panel-body">
			<div class="micro">Most recent status changes recorded in <span class="mono">state_transitions</span>.</div>
			<div class="action-lines" id="transition-lines" style="margin-top:14px;"></div>
          </div>
        </div>

      </div>
    </div>

    <div class="grid metrics" id="metrics" style="margin-top:16px;"></div>
  </div>

  <script>
    const state = { snapshot: null };

    const fmtMoney = cents => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format((Number(cents || 0) / 100));
    const fmtInt = value => new Intl.NumberFormat('en-US').format(Number(value || 0));
    const fmtPct = value => Number(value || 0).toFixed(1) + '%';
    const toneClass = tone => ({ success: 'success-badge', info: 'info', warning: 'warning', danger: 'danger', accent: 'info' }[tone] || 'neutral');

		function transitionCategory(state) {
			return { row: 'transition-gray', badge: 'neutral' };
		}

    function setAction(report) {
      if (!report) return;
      document.getElementById('hero-title').textContent = report.message || 'Action complete.';
      document.getElementById('hero-sub').textContent = report.error ? 'The control plane reported an error. Check the action panel.' : 'The feed is now watching for new DB activity and processor progress.';
      document.getElementById('feed-badge').textContent = report.error ? 'needs attention' : 'live';
      document.getElementById('feed-badge').className = 'badge ' + (report.error ? 'danger' : 'success-badge');
    }


		function renderStateTransitions(snapshot) {
			const root = document.getElementById('transition-lines');
			root.innerHTML = '';
			const items = snapshot.state_transitions || [];
			if (!items.length) {
				root.innerHTML = '<div class="small">No state transitions yet.</div>';
				return;
			}
			items.forEach(item => {
				const row = document.createElement('div');
				const category = transitionCategory(item.to_state);
				row.className = 'action-line transition-line ' + category.row;
				const fromState = item.from_state ? item.from_state : 'START';
				const when = new Date(item.transitioned_at).toLocaleTimeString([], {hour:'2-digit', minute:'2-digit', second:'2-digit'});
				const reason = item.reason ? ('<div class="small" style="margin-top:4px;">' + item.reason + '</div>') : '';
				row.innerHTML = '<div><div class="bold">' + item.entity_type + ' • ' + item.entity_id + '</div><div class="small">' + fromState + ' -> ' + item.to_state + ' at ' + when + '</div>' + reason + '</div><span class="badge ' + category.badge + '">' + item.to_state + '</span>';
				root.appendChild(row);
			});
		}


    function renderFeed(items) {
      const root = document.getElementById('feed');
      root.innerHTML = '';
      if (!items.length) {
        root.innerHTML = '<div class="small">No live events yet. Use the buttons above to start the flow.</div>';
        return;
      }
      items.slice().reverse().forEach(item => {
        const row = document.createElement('div');
        row.className = 'feed-item';
				const body = item.message ? '<div class="body">' + item.message + '</div>' : '';
				row.innerHTML = '<div class="meta">' + new Date(item.at).toLocaleTimeString([], {hour:'2-digit', minute:'2-digit', second:'2-digit'}) + '</div><div><div class="title">' + item.title + '</div>' + body + '</div><span class="badge ' + toneClass(item.tone) + ' status">' + (item.status || item.kind) + '</span>';
        root.appendChild(row);
      });
    }

    async function refreshSnapshot() {
      const res = await fetch('/api/snapshot');
      const snapshot = await res.json();
      state.snapshot = snapshot;
			renderStateTransitions(snapshot);
      renderFeed(snapshot.live_feed || []);
      if (snapshot.last_action) setAction(snapshot.last_action);
    }

    async function invoke(path) {
      const res = await fetch(path, { method: 'POST' });
      const data = await res.json();
      setAction(data);
      await refreshSnapshot();
    }

		document.getElementById('ingest-btn').addEventListener('click', () => invoke('/api/ingest-reconcile').catch(console.error));
		const refreshButton = document.getElementById('refresh-btn');
		if (refreshButton) {
			refreshButton.addEventListener('click', () => refreshSnapshot().catch(console.error));
		}

    const source = new EventSource('/events');
    source.onopen = () => {
      document.getElementById('feed-badge').textContent = 'live';
      document.getElementById('feed-badge').className = 'badge success-badge';
    };
    source.onmessage = event => {
      try {
        const payload = JSON.parse(event.data);
        const feed = (state.snapshot && state.snapshot.live_feed) ? state.snapshot.live_feed.slice() : [];
        feed.push(payload);
        if (!state.snapshot) state.snapshot = {};
        state.snapshot.live_feed = feed.slice(-20);
        renderFeed(state.snapshot.live_feed);
      } catch (err) {
        console.error(err);
      }
    };
    source.onerror = () => {
      document.getElementById('feed-badge').textContent = 'reconnecting';
      document.getElementById('feed-badge').className = 'badge warning';
    };

    refreshSnapshot().catch(console.error);
    setInterval(() => refreshSnapshot().catch(console.error), 5000);
  </script>
</body>
</html>`

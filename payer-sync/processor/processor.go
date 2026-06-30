package processor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	retry "github.com/sethvargo/go-retry"

	"weavelab.xyz/payer-sync/internal/db/store"
)

const reconciledPaymentEntityType = "reconciled_payment"

type PaymentStatus string

const (
	StatusMatched           PaymentStatus = "MATCHED"
	StatusProcessingPayment PaymentStatus = "PROCESSING_PAYMENT"
	StatusPaymentSucceeded  PaymentStatus = "PAYMENT_SUCCEEDED"
	StatusProcessingFailed  PaymentStatus = "PROCESSING_FAILED"
)

func (s PaymentStatus) String() string { return string(s) }

type txStarter interface {
	store.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PaymentProcessor is the interface for charging a virtual credit card.
type PaymentProcessor interface {
	CreatePaymentMethod(ctx context.Context, req CreatePaymentMethodRequest) (*PaymentMethod, error)
	CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*PaymentIntent, error)
	ConfirmPaymentIntent(ctx context.Context, paymentIntentID, idempotencyKey string) (*PaymentIntent, error)
}

type CreatePaymentMethodRequest struct {
	CardNumber string
	ExpMonth   string
	ExpYear    string
	CVV        string
}

type CreatePaymentIntentRequest struct {
	AmountCents     int64
	Currency        string
	PaymentMethodID string
	IdempotencyKey  string
	Metadata        map[string]string
}

type PaymentMethod struct {
	ID    string
	Last4 string
}

type PaymentIntent struct {
	ID       string
	Status   string
	ChargeID string
}

type ProcessorError struct {
	Code    string
	Message string
}

func (e *ProcessorError) Error() string { return e.Message }

func (e *ProcessorError) IsRetryable() bool {
	switch e.Code {
	case "processor_unavailable", "network_failure", "rate_limit":
		return true
	default:
		return false
	}
}

type Config struct {
	Now        func() time.Time
	NewID      func(prefix string) string
	MaxRetries int
	Processor  PaymentProcessor
	ListenConn *pgx.Conn
}

type Service struct {
	db         txStarter
	q          *store.Queries
	now        func() time.Time
	newID      func(prefix string) string
	maxRetries int
	processor  PaymentProcessor
	listenConn *pgx.Conn // dedicated connection for LISTEN only; never used for queries
}

func NewService(db txStarter, cfg Config) *Service {
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	newIDFn := cfg.NewID
	if newIDFn == nil {
		newIDFn = defaultID
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &Service{
		db:         db,
		q:          store.New(db),
		now:        nowFn,
		newID:      newIDFn,
		maxRetries: maxRetries,
		processor:  cfg.Processor,
		listenConn: cfg.ListenConn,
	}
}

// Run listens for notifications and drains any already-matched backlog.
// It returns nil when ctx is cancelled (clean shutdown) and an error on unexpected failures.
func (s *Service) Run(ctx context.Context) error {
	if s.listenConn == nil {
		return fmt.Errorf("listen connection is nil")
	}

	// Establish LISTEN before draining so notifications are queued while we process backlog.
	if _, err := s.listenConn.Exec(ctx, "LISTEN reconciled_payment_matched"); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("listen: %w", err)
	}
	slog.Info("processor listening for notifications", "channel", "reconciled_payment_matched")

	// Drain pass: process all currently MATCHED payments after LISTEN is active.
	payments, err := s.q.ListMatchedPaymentsForProcessing(ctx, 1000)
	if err != nil {
		return fmt.Errorf("drain pass: %w", err)
	}
	slog.Info("processor drain pass complete", "matched_payments", len(payments))
	for _, p := range payments {
		if err := s.processPayment(ctx, p.ReconciledPaymentID); err != nil {
			return fmt.Errorf("drain process payment %s: %w", p.ReconciledPaymentID, err)
		}
	}

	for {
		notification, err := s.listenConn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("wait for notification: %w", err)
		}
		slog.Info("processor notification received", "payment_id", notification.Payload)
		if err := s.processPayment(ctx, notification.Payload); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("process payment %s: %w", notification.Payload, err)
		}
	}
}

func (s *Service) processPayment(ctx context.Context, id string) error {
	now := s.now().UTC()

	payment, err := s.q.GetMatchedPaymentWithVCCDetails(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("get payment %s: %w", id, err)
	}

	idempotencyKey := payment.IdempotencyKey.String
	if strings.TrimSpace(idempotencyKey) == "" {
		idempotencyKey = computeIdempotencyKey(id)
	}

	// BeginProcessing is the only state gate — optimistic lock via AND status = 'MATCHED'.
	_, err = s.q.BeginProcessing(ctx, store.BeginProcessingParams{
		ReconciledPaymentID: id,
		IdempotencyKey:      pgtype.Text{String: idempotencyKey, Valid: true},
		ProcessingStartedAt: timestamptz(now),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Info("processor payment already claimed or not matched", "payment_id", id)
			return nil // already claimed or not in MATCHED state
		}
		return fmt.Errorf("begin processing %s: %w", id, err)
	}
	slog.Info("processor claimed payment",
		"payment_id", id,
		"trace_number", payment.TraceNumber,
		"amount", numericToString(payment.MatchedAmount),
	)

	if err := s.insertTransition(ctx, id, StatusMatched, StatusProcessingPayment, "processor claimed payment"); err != nil {
		return err
	}

	attemptNum := 0
	maxRetries := s.maxRetries
	// WithMaxRetries(n) allows n retries beyond the first attempt; total = n+1.
	// We want maxRetries total attempts, so pass maxRetries-1.
	backoff := retry.WithMaxRetries(uint64(maxRetries-1), retry.NewExponential(500*time.Millisecond))

	retryErr := retry.Do(ctx, backoff, func(ctx context.Context) error {
		attemptNum++
		attemptNow := s.now().UTC()

		pmID := payment.PaymentMethodID.String
		if pmID == "" {
			err := &ProcessorError{Code: "missing_payment_method", Message: "payment has no tokenized payment method ID"}
			_ = s.insertAttempt(ctx, id, idempotencyKey, attemptNum, "failed", err, attemptNow)
			return err
		}

		amountCents, err := numericToCents(payment.MatchedAmount)
		if err != nil {
			return fmt.Errorf("compute amount cents: %w", err)
		}

		pi, err := s.processor.CreatePaymentIntent(ctx, CreatePaymentIntentRequest{
			AmountCents:     amountCents,
			Currency:        "usd",
			PaymentMethodID: pmID,
			IdempotencyKey:  idempotencyKey,
			Metadata: map[string]string{
				"reconciled_payment_id": id,
				"location_id":           payment.LocationID,
				"trace_number":          payment.TraceNumber,
				"era_id":                payment.EraPaymentGroupID,
				"vcc_payment_group_id":  payment.VccPaymentGroupID,
				"payer_name":            payment.PayerName.String,
			},
		})
		if err != nil {
			procErr := toProcessorError(err)
			outcome := "retrying"
			if !procErr.IsRetryable() {
				outcome = "failed"
			}
			_ = s.insertAttempt(ctx, id, idempotencyKey, attemptNum, outcome, procErr, attemptNow)
			if procErr.IsRetryable() {
				return retry.RetryableError(procErr)
			}
			return procErr
		}

		confirmed, err := s.processor.ConfirmPaymentIntent(ctx, pi.ID, idempotencyKey)
		if err != nil {
			procErr := toProcessorError(err)
			outcome := "retrying"
			if !procErr.IsRetryable() {
				outcome = "failed"
			}
			_ = s.insertAttempt(ctx, id, idempotencyKey, attemptNum, outcome, procErr, attemptNow)
			if procErr.IsRetryable() {
				return retry.RetryableError(procErr)
			}
			return procErr
		}

		// Success path — PAYMENT_SUCCEEDED is only set after a confirmed PaymentIntent.
		_ = s.insertAttempt(ctx, id, idempotencyKey, attemptNum, "succeeded", nil, attemptNow)

		_, err = s.q.MarkPaymentSucceeded(ctx, store.MarkPaymentSucceededParams{
			ReconciledPaymentID:      id,
			ProcessorPaymentIntentID: pgtype.Text{String: confirmed.ID, Valid: true},
			ProcessingCompletedAt:    timestamptz(attemptNow),
		})
		if err != nil {
			return fmt.Errorf("mark succeeded: %w", err)
		}

		slog.Info("processor payment succeeded", "payment_id", id, "intent_id", confirmed.ID)
		return s.insertTransition(ctx, id, StatusProcessingPayment, StatusPaymentSucceeded, "payment confirmed by processor")
	})

	if retryErr == nil {
		return nil
	}

	// Terminal failure or retry exhaustion.
	failNow := s.now().UTC()
	if ctx.Err() != nil {
		if err := s.requeueClaimedPayment(id, failNow); err != nil {
			return err
		}
		slog.Info("processor payment requeued after shutdown", "payment_id", id)
		return nil
	}

	var procErr *ProcessorError
	errCode := ""
	errMsg := retryErr.Error()
	if errors.As(retryErr, &procErr) {
		errCode = procErr.Code
		errMsg = procErr.Message
	}

	if _, err := s.q.MarkPaymentFailed(ctx, store.MarkPaymentFailedParams{
		ReconciledPaymentID:   id,
		ProcessorErrorCode:    pgtype.Text{String: errCode, Valid: errCode != ""},
		ProcessorErrorMessage: pgtype.Text{String: errMsg, Valid: errMsg != ""},
		RetryCount:            int32(max(attemptNum-1, 0)),
		ProcessingCompletedAt: timestamptz(failNow),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("processor payment no longer exists while marking failed",
				"payment_id", id,
				"attempts", attemptNum,
				"code", errCode,
				"message", errMsg,
			)
			return nil
		}
		return fmt.Errorf("mark failed: %w", err)
	}

	slog.Error("processor payment failed",
		"payment_id", id,
		"attempts", attemptNum,
		"code", errCode,
		"message", errMsg,
	)
	return s.insertTransition(ctx, id, StatusProcessingPayment, StatusProcessingFailed, errMsg)
}

func (s *Service) requeueClaimedPayment(id string, at time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.q.RequeueClaimedPayment(ctx, store.RequeueClaimedPaymentParams{
		ReconciledPaymentID: id,
		UpdatedAt:           timestamptz(at),
	})
	if err != nil {
		return fmt.Errorf("requeue payment %s: %w", id, err)
	}
	if rows == 0 {
		slog.Info("processor requeue skipped; payment not in processing state", "payment_id", id)
	}
	return nil
}

func (s *Service) insertAttempt(ctx context.Context, paymentID, idempotencyKey string, attemptNum int, outcome string, procErr *ProcessorError, at time.Time) error {
	var errCode, errMsg pgtype.Text
	if procErr != nil {
		errCode = pgtype.Text{String: procErr.Code, Valid: procErr.Code != ""}
		errMsg = pgtype.Text{String: procErr.Message, Valid: procErr.Message != ""}
	}
	_, err := s.q.InsertProcessorAttempt(ctx, store.InsertProcessorAttemptParams{
		AttemptID:           s.newID("attempt"),
		ReconciledPaymentID: paymentID,
		IdempotencyKey:      idempotencyKey,
		AttemptNumber:       int32(attemptNum),
		Outcome:             outcome,
		ErrorCode:           errCode,
		ErrorMessage:        errMsg,
		AttemptedAt:         timestamptz(at),
	})
	return err
}

func (s *Service) insertTransition(ctx context.Context, entityID string, fromState, toState PaymentStatus, reason string) error {
	_, err := s.q.InsertProcessorStateTransition(ctx, store.InsertProcessorStateTransitionParams{
		TransitionID:   s.newID("transition"),
		EntityType:     reconciledPaymentEntityType,
		EntityID:       entityID,
		FromState:      pgtype.Text{String: fromState.String(), Valid: fromState != ""},
		ToState:        toState.String(),
		TransitionedAt: timestamptz(s.now().UTC()),
		Reason:         pgtype.Text{String: reason, Valid: reason != ""},
	})
	if err != nil {
		return fmt.Errorf("insert transition %s→%s for %s: %w", fromState, toState, entityID, err)
	}
	return nil
}

// computeIdempotencyKey generates a deterministic key for a single reconciled payment.
// Same payment ID always produces the same key.
func computeIdempotencyKey(reconciledPaymentID string) string {
	raw := fmt.Sprintf("reconciled-payment:%s", reconciledPaymentID)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func toProcessorError(err error) *ProcessorError {
	var pe *ProcessorError
	if errors.As(err, &pe) {
		return pe
	}
	return &ProcessorError{Code: "network_failure", Message: err.Error()}
}

func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0"
	}
	r := numericToRat(n)
	if r == nil {
		return "0"
	}
	return r.FloatString(2)
}

func numericToCents(n pgtype.Numeric) (int64, error) {
	if !n.Valid {
		return 0, fmt.Errorf("null amount")
	}
	r := numericToRat(n)
	if r == nil {
		return 0, fmt.Errorf("invalid amount")
	}
	r.Mul(r, new(big.Rat).SetInt64(100))
	if !r.IsInt() {
		return 0, fmt.Errorf("amount has sub-cent precision")
	}
	if !r.Num().IsInt64() {
		return 0, fmt.Errorf("amount exceeds int64 range")
	}
	return r.Num().Int64(), nil
}

func numericToRat(n pgtype.Numeric) *big.Rat {
	if !n.Valid || n.NaN {
		return nil
	}
	r := new(big.Rat)
	if n.Int != nil {
		r.SetInt(n.Int)
	}
	if n.Exp > 0 {
		exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)
		r.Mul(r, new(big.Rat).SetInt(exp))
	} else if n.Exp < 0 {
		exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)
		r.Quo(r, new(big.Rat).SetInt(exp))
	}
	return r
}

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func defaultID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}

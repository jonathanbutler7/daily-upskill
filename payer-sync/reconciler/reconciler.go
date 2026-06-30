package reconciler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"weavelab.xyz/payer-sync/internal/db/store"
)

const (
	eraAwaitingStatus           = "AWAITING_VCC"
	vccAwaitingStatus           = "AWAITING_ERA"
	matchedStatus               = "MATCHED"
	exceptionUnmatchedStatus    = "EXCEPTION_UNMATCHED"
	reconciledPaymentEntityType = "reconciled_payment"
	eraPaymentGroupEntityType   = "era_payment_group"
	vccPaymentGroupEntityType   = "vcc_payment_group"
	expirationReason            = "unmatched for more than 5 business days"
)

type txStarter interface {
	store.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Alerter is notified when payment groups expire unmatched so that operations
// can investigate. Implementations should be non-blocking or handle their own
// retries; a failed alert does not abort the expiration.
type Alerter interface {
	AlertUnmatched(ctx context.Context, entityType, groupID, locationID string) error
}

type Config struct {
	Now     func() time.Time
	NewID   func(prefix string) string
	Alerter Alerter
	Logger  *slog.Logger
}

type Service struct {
	db      txStarter
	q       *store.Queries
	now     func() time.Time
	newID   func(prefix string) string
	alerter Alerter
	log     *slog.Logger
}

type RunResult struct {
	RunID           string
	MatchedCount    int
	ExpiredERACount int
	ExpiredVCCCount int
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

	return &Service{
		db:      db,
		q:       store.New(db),
		now:     nowFn,
		newID:   newIDFn,
		alerter: cfg.Alerter,
		log:     loggerOrDefault(cfg.Logger),
	}
}

func (s *Service) Run(ctx context.Context) (result RunResult, runErr error) {
	now := s.now().UTC()
	result.RunID = s.newID("reconciler-run")
	s.log.Info("reconciler run started", "run_id", result.RunID)

	if err := s.createJobRun(ctx, result.RunID, now); err != nil {
		return result, err
	}

	defer func() {
		if err := s.finishJobRun(ctx, result, s.now().UTC(), runErr); err != nil && runErr == nil {
			runErr = err
		}
	}()

	matchedCount, err := s.reconcileMatches(ctx, now)
	if err != nil {
		return result, err
	}
	result.MatchedCount = matchedCount
	s.log.Info("reconciler matches complete", "run_id", result.RunID, "matched", matchedCount)

	expiredERAs, err := s.expireStaleERAGroups(ctx, now)
	if err != nil {
		return result, err
	}
	result.ExpiredERACount = expiredERAs
	s.log.Info("reconciler era expirations complete", "run_id", result.RunID, "expired_era", expiredERAs)

	expiredVCCs, err := s.expireStaleVCCGroups(ctx, now)
	if err != nil {
		return result, err
	}
	result.ExpiredVCCCount = expiredVCCs
	s.log.Info("reconciler vcc expirations complete", "run_id", result.RunID, "expired_vcc", expiredVCCs)

	s.log.Info("reconciler run complete",
		"run_id", result.RunID,
		"matched", result.MatchedCount,
		"expired_era", result.ExpiredERACount,
		"expired_vcc", result.ExpiredVCCCount,
	)

	return result, nil
}

func (s *Service) reconcileMatches(ctx context.Context, now time.Time) (int, error) {
	eras, err := s.q.ListUnmatchedERAPaymentGroups(ctx)
	if err != nil {
		return 0, fmt.Errorf("list unmatched ERA groups: %w", err)
	}
	vccGroups, err := s.q.ListUnmatchedVCCPaymentGroups(ctx)
	if err != nil {
		return 0, fmt.Errorf("list unmatched VCC groups: %w", err)
	}

	vccByKey := make(map[string][]store.ListUnmatchedVCCPaymentGroupsRow, len(vccGroups))
	for _, vcc := range vccGroups {
		key := counterpartKey(vcc.LocationID, vcc.TraceID)
		vccByKey[key] = append(vccByKey[key], vcc)
	}

	matched := 0
	for _, era := range eras {
		candidates := vccByKey[counterpartKey(era.LocationID, era.TraceNumber)]

		candidate, ok := selectMatchCandidate(era, candidates)
		if !ok {
			continue
		}

		persisted, err := s.persistMatch(ctx, now, era, candidate)
		if err != nil {
			return matched, err
		}
		if persisted {
			matched++
		}
	}

	return matched, nil
}

func counterpartKey(locationID, traceID string) string {
	return locationID + "\x00" + traceID
}

func selectMatchCandidate(era store.ListUnmatchedERAPaymentGroupsRow, candidates []store.ListUnmatchedVCCPaymentGroupsRow) (store.ListUnmatchedVCCPaymentGroupsRow, bool) {
	for _, candidate := range candidates {
		if !numericStringEqual(era.BprAmount, candidate.TotalAmount) {
			continue
		}
		if !providerIdentityConsistent(era.ProviderNpi.String, era.ProviderTaxID.String, candidate.ProviderNpi.String, candidate.ProviderTaxID.String) {
			continue
		}
		return candidate, true
	}

	return store.ListUnmatchedVCCPaymentGroupsRow{}, false
}

func providerIdentityConsistent(eraNPI, eraTaxID, vccNPI, vccTaxID string) bool {
	eraNPI = strings.TrimSpace(eraNPI)
	eraTaxID = strings.TrimSpace(eraTaxID)
	vccNPI = strings.TrimSpace(vccNPI)
	vccTaxID = strings.TrimSpace(vccTaxID)

	matched := false

	if eraNPI != "" && vccNPI != "" {
		if eraNPI != vccNPI {
			return false
		}
		matched = true
	}

	if eraTaxID != "" && vccTaxID != "" {
		if eraTaxID != vccTaxID {
			return false
		}
		matched = true
	}

	return matched
}

func (s *Service) persistMatch(ctx context.Context, now time.Time, era store.ListUnmatchedERAPaymentGroupsRow, vcc store.ListUnmatchedVCCPaymentGroupsRow) (bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin match tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := s.q.WithTx(tx)
	matchedAmount, err := numericFromString(era.BprAmount)
	if err != nil {
		return false, fmt.Errorf("parse matched amount for ERA %s: %w", era.GroupID, err)
	}

	reconciledPaymentID := s.newID("reconciled-payment")
	if _, err := qtx.CreateReconciledPayment(ctx, store.CreateReconciledPaymentParams{
		ReconciledPaymentID: reconciledPaymentID,
		LocationID:          era.LocationID,
		EraPaymentGroupID:   era.GroupID,
		VccPaymentGroupID:   vcc.GroupID,
		TraceNumber:         era.TraceNumber,
		MatchedAmount:       matchedAmount,
		PayerName:           era.PayerName,
		ProviderNpi:         preferredProviderID(era.ProviderNpi, vcc.ProviderNpi),
		ProviderTaxID:       preferredProviderID(era.ProviderTaxID, vcc.ProviderTaxID),
		Status:              matchedStatus,
		MatchedAt:           timestamptz(now),
	}); err != nil {
		if isUniqueViolation(err) {
			s.log.Info("reconciler match skipped due to concurrent insert",
				"era_group_id", era.GroupID,
				"vcc_group_id", vcc.GroupID,
			)
			return false, nil
		}
		return false, fmt.Errorf("create reconciled payment for ERA %s and VCC %s: %w", era.GroupID, vcc.GroupID, err)
	}

	// Notify the processor only if this transaction commits.
	if err := qtx.NotifyReconciledPaymentMatched(ctx, reconciledPaymentID); err != nil {
		return false, fmt.Errorf("notify processor for reconciled payment %s: %w", reconciledPaymentID, err)
	}

	if _, err := qtx.MarkERAPaymentGroupMatched(ctx, store.MarkERAPaymentGroupMatchedParams{
		GroupID:   era.GroupID,
		MatchedAt: timestamptz(now),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Info("reconciler match skipped because ERA group is no longer awaiting",
				"era_group_id", era.GroupID,
			)
			return false, nil
		}
		return false, fmt.Errorf("mark ERA %s matched: %w", era.GroupID, err)
	}

	if _, err := qtx.MarkVCCPaymentGroupMatched(ctx, store.MarkVCCPaymentGroupMatchedParams{
		GroupID:   vcc.GroupID,
		MatchedAt: timestamptz(now),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Info("reconciler match skipped because VCC group is no longer awaiting",
				"vcc_group_id", vcc.GroupID,
			)
			return false, nil
		}
		return false, fmt.Errorf("mark VCC %s matched: %w", vcc.GroupID, err)
	}

	if err := insertTransition(ctx, qtx, s.newID, now, eraPaymentGroupEntityType, era.GroupID, era.Status, matchedStatus, "matched to VCC payment group"); err != nil {
		return false, err
	}
	if err := insertTransition(ctx, qtx, s.newID, now, vccPaymentGroupEntityType, vcc.GroupID, vcc.Status, matchedStatus, "matched to ERA payment group"); err != nil {
		return false, err
	}
	if err := insertTransition(ctx, qtx, s.newID, now, reconciledPaymentEntityType, reconciledPaymentID, "", matchedStatus, "reconciled payment created"); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit match tx: %w", err)
	}

	return true, nil
}

func preferredProviderID(primary, fallback pgtype.Text) pgtype.Text {
	if primary.Valid && strings.TrimSpace(primary.String) != "" {
		return primary
	}
	return fallback
}

func (s *Service) expireStaleERAGroups(ctx context.Context, now time.Time) (int, error) {
	eras, err := s.q.ListUnmatchedERAPaymentGroups(ctx)
	if err != nil {
		return 0, fmt.Errorf("list unmatched ERA groups for expiration: %w", err)
	}

	expired := 0
	for _, era := range eras {
		if !shouldExpire(era.FirstReceivedAt.Time, now) {
			continue
		}
		if err := s.expireERAGroup(ctx, now, era); err != nil {
			return expired, err
		}
		expired++
	}

	return expired, nil
}

func (s *Service) expireStaleVCCGroups(ctx context.Context, now time.Time) (int, error) {
	vccGroups, err := s.q.ListUnmatchedVCCPaymentGroups(ctx)
	if err != nil {
		return 0, fmt.Errorf("list unmatched VCC groups for expiration: %w", err)
	}

	expired := 0
	for _, vcc := range vccGroups {
		if !shouldExpire(vcc.FirstReceivedAt.Time, now) {
			continue
		}
		if err := s.expireVCCGroup(ctx, now, vcc); err != nil {
			return expired, err
		}
		expired++
	}

	return expired, nil
}

func shouldExpire(firstReceivedAt, now time.Time) bool {
	deadline := addBusinessDays(firstReceivedAt.UTC(), 5)
	return now.UTC().After(deadline)
}

// Business days in MVP are weekdays only; holidays are intentionally ignored.
func addBusinessDays(start time.Time, businessDays int) time.Time {
	current := start
	added := 0
	for added < businessDays {
		current = current.AddDate(0, 0, 1)
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			continue
		}
		added++
	}
	return current
}

func (s *Service) expireERAGroup(ctx context.Context, now time.Time, era store.ListUnmatchedERAPaymentGroupsRow) error {
	groupID := era.GroupID
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin ERA expiration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := s.q.WithTx(tx)
	row, err := qtx.ExpireUnmatchedERAPaymentGroup(ctx, store.ExpireUnmatchedERAPaymentGroupParams{
		GroupID:         groupID,
		ExceptionAt:     timestamptz(now),
		ExceptionReason: text(expirationReason),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Info("reconciler era expiration skipped because group is no longer awaiting", "era_group_id", groupID)
			return nil
		}
		return fmt.Errorf("expire ERA group %s: %w", groupID, err)
	}

	if err := insertTransition(ctx, qtx, s.newID, now, eraPaymentGroupEntityType, row.GroupID, row.PriorStatus.String, row.Status, expirationReason); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ERA expiration tx: %w", err)
	}

	if s.alerter != nil {
		if err := s.alerter.AlertUnmatched(ctx, eraPaymentGroupEntityType, groupID, era.LocationID); err != nil {
			return fmt.Errorf("alert unmatched ERA group %s: %w", groupID, err)
		}
	}
	return nil
}

func (s *Service) expireVCCGroup(ctx context.Context, now time.Time, vcc store.ListUnmatchedVCCPaymentGroupsRow) error {
	groupID := vcc.GroupID
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin VCC expiration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := s.q.WithTx(tx)
	row, err := qtx.ExpireUnmatchedVCCPaymentGroup(ctx, store.ExpireUnmatchedVCCPaymentGroupParams{
		GroupID:         groupID,
		ExceptionAt:     timestamptz(now),
		ExceptionReason: text(expirationReason),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Info("reconciler vcc expiration skipped because group is no longer awaiting", "vcc_group_id", groupID)
			return nil
		}
		return fmt.Errorf("expire VCC group %s: %w", groupID, err)
	}

	if err := insertTransition(ctx, qtx, s.newID, now, vccPaymentGroupEntityType, row.GroupID, row.PriorStatus.String, row.Status, expirationReason); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit VCC expiration tx: %w", err)
	}

	if s.alerter != nil {
		if err := s.alerter.AlertUnmatched(ctx, vccPaymentGroupEntityType, groupID, vcc.LocationID); err != nil {
			return fmt.Errorf("alert unmatched VCC group %s: %w", groupID, err)
		}
	}
	return nil
}

func insertTransition(ctx context.Context, qtx *store.Queries, newID func(string) string, now time.Time, entityType, entityID, fromState, toState, reason string) error {
	_, err := qtx.InsertReconcilerStateTransition(ctx, store.InsertReconcilerStateTransitionParams{
		TransitionID:   newID("transition"),
		EntityType:     entityType,
		EntityID:       entityID,
		FromState:      text(fromState),
		ToState:        toState,
		TransitionedAt: timestamptz(now),
		Reason:         text(reason),
	})
	if err != nil {
		return fmt.Errorf("insert transition for %s %s: %w", entityType, entityID, err)
	}
	return nil
}

func (s *Service) createJobRun(ctx context.Context, runID string, startedAt time.Time) error {
	_, err := s.q.CreateReconcilerJobRun(ctx, store.CreateReconcilerJobRunParams{
		RunID:          runID,
		StartedAt:      timestamptz(startedAt),
		FinishedAt:     pgtype.Timestamptz{},
		Status:         "running",
		FilesProcessed: 0,
		RecordsMatched: 0,
		Errors:         []byte("[]"),
	})
	if err != nil {
		return fmt.Errorf("create job run %s: %w", runID, err)
	}
	return nil
}

func (s *Service) finishJobRun(ctx context.Context, result RunResult, finishedAt time.Time, runErr error) error {
	status := "success"
	var payload []string
	if runErr != nil {
		status = "failure"
		payload = append(payload, runErr.Error())
	}

	rawErrors, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal job run errors: %w", err)
	}

	if err := s.q.UpdateReconcilerJobRunResult(ctx, store.UpdateReconcilerJobRunResultParams{
		RunID:          result.RunID,
		FinishedAt:     timestamptz(finishedAt),
		Status:         status,
		FilesProcessed: 0,
		RecordsMatched: int32(result.MatchedCount),
		Errors:         rawErrors,
	}); err != nil {
		return fmt.Errorf("update job run %s: %w", result.RunID, err)
	}
	return nil
}

func text(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: strings.TrimSpace(value) != ""}
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func numericFromString(value string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(value); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

// numericStringEqual parses both values as NUMERIC and compares them
// mathematically, so "100.00" and "100.0" are considered equal.
func numericStringEqual(a, b string) bool {
	var na, nb pgtype.Numeric
	if err := na.Scan(strings.TrimSpace(a)); err != nil {
		return false
	}
	if err := nb.Scan(strings.TrimSpace(b)); err != nil {
		return false
	}
	ra := numericToRat(na)
	rb := numericToRat(nb)
	if ra == nil || rb == nil {
		return false
	}
	return ra.Cmp(rb) == 0
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

func defaultID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}

func loggerOrDefault(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.Default()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

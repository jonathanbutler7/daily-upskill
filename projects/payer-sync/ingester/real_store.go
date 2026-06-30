package ingester

import (
	"context"
	"errors"
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"weavelab.xyz/payer-sync/internal/db/store"
)

// RealIngesterStore is the production implementation of IngesterStore.
type RealIngesterStore struct {
	q *store.Queries
}

func NewRealIngesterStore(db store.DBTX) *RealIngesterStore {
	return &RealIngesterStore{q: store.New(db)}
}

// ---- ERA remittances ----

func (s *RealIngesterStore) ExistsERAByOfficeAndHash(ctx context.Context, locationID, fileHash string) (bool, error) {
	_, err := s.q.GetERARemittanceByLocationAndHash(ctx, store.GetERARemittanceByLocationAndHashParams{
		LocationID: locationID, FileHash: fileHash,
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (s *RealIngesterStore) CreateERA(ctx context.Context, p CreateERAParams) error {
	bprNum, err := decimalStringToNumeric(p.BPRAmount)
	if err != nil {
		return fmt.Errorf("store CreateERA: invalid BPRAmount %q: %w", p.BPRAmount, err)
	}
	_, err = s.q.CreateERARemittance(ctx, store.CreateERARemittanceParams{
		EraID:         p.ID,
		LocationID:    p.LocationID,
		PayerName:     pgText(p.PayerName),
		ProviderNpi:   pgText(p.ProviderNPI),
		ProviderTaxID: pgText(p.ProviderTaxID),
		BprAmount:     bprNum,
		PaymentMethod: pgText(p.PaymentMethod),
		TraceNumber:   pgText(p.TraceNumber),
		Status:        p.Status,
		ReceivedAt:    pgNow(),
		FileHash:      p.FileHash,
		RawStorageKey: p.StorageKey,
	})
	return err
}

func (s *RealIngesterStore) UpdateERAParsed(ctx context.Context, p UpdateERAParsedParams) error {
	bprNum, err := decimalStringToNumeric(p.BPRAmount)
	if err != nil {
		return fmt.Errorf("store UpdateERAParsed: invalid BPRAmount %q: %w", p.BPRAmount, err)
	}
	return s.q.UpdateERARemittanceParsed(ctx, store.UpdateERARemittanceParsedParams{
		EraID:         p.ID,
		PayerName:     pgText(p.PayerName),
		ProviderNpi:   pgText(p.ProviderNPI),
		ProviderTaxID: pgText(p.ProviderTaxID),
		BprAmount:     bprNum,
		PaymentMethod: pgText(p.PaymentMethod),
		TraceNumber:   pgText(p.TraceNumber),
	})
}

func (s *RealIngesterStore) SetERAParseFailure(ctx context.Context, eraID string) error {
	return s.q.SetERARemittanceParseFailure(ctx, eraID)
}

// ---- VCC files ----

func (s *RealIngesterStore) ExistsVCCByOfficeAndHash(ctx context.Context, locationID, fileHash string) (bool, error) {
	_, err := s.q.GetVCCFileByLocationAndHash(ctx, store.GetVCCFileByLocationAndHashParams{
		LocationID: locationID, FileHash: fileHash,
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (s *RealIngesterStore) CreateVCCFile(ctx context.Context, p CreateVCCFileParams) error {
	_, err := s.q.CreateVCCFile(ctx, store.CreateVCCFileParams{
		VccFileID:      p.ID,
		LocationID:     p.LocationID,
		ReceivedAt:     pgNow(),
		FileHash:       p.FileHash,
		RawStorageKey:  p.StorageKey,
		RowCount:       int32(p.RowCount),
		SourceFilename: p.SourceFilename,
		Status:         p.Status,
	})
	return err
}

func (s *RealIngesterStore) UpdateVCCFileParsed(ctx context.Context, p UpdateVCCFileParsedParams) error {
	return s.q.UpdateVCCFileParsed(ctx, store.UpdateVCCFileParsedParams{
		VccFileID: p.ID,
		RowCount:  int32(p.RowCount),
	})
}

func (s *RealIngesterStore) SetVCCFileParseFailure(ctx context.Context, vccFileID string, rowCount int) error {
	return s.q.SetVCCFileParseFailure(ctx, store.SetVCCFileParseFailureParams{
		VccFileID: vccFileID,
		RowCount:  int32(rowCount),
	})
}

// ---- ERA payment groups ----

func (s *RealIngesterStore) GetActiveERAPaymentGroup(ctx context.Context, locationID, traceNumber string) (*ERAPaymentGroup, error) {
	row, err := s.q.GetActiveERAPaymentGroup(ctx, store.GetActiveERAPaymentGroupParams{
		LocationID: locationID, TraceNumber: traceNumber,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return eraRowToGroup(row.GroupID, row.EraID, row.LocationID, row.TraceNumber, row.BprAmount,
		row.ClaimCount, row.Claims, row.Adjustments, row.Status, row.ReconciliationTriggeredAt, row.CreatedAt), nil
}

func (s *RealIngesterStore) GetMatchingERAForVCC(ctx context.Context, locationID, traceID string) (*ERAPaymentGroup, error) {
	row, err := s.q.GetMatchingERAForVCC(ctx, store.GetMatchingERAForVCCParams{
		LocationID: locationID, TraceNumber: traceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return eraRowToGroup(row.GroupID, row.EraID, row.LocationID, row.TraceNumber, row.BprAmount,
		row.ClaimCount, row.Claims, row.Adjustments, row.Status, row.ReconciliationTriggeredAt, row.CreatedAt), nil
}

func (s *RealIngesterStore) CreateERAPaymentGroup(ctx context.Context, g *ERAPaymentGroup) error {
	bprNum, err := decimalStringToNumeric(g.BPRAmount)
	if err != nil {
		return fmt.Errorf("store CreateERAPaymentGroup: invalid BPRAmount %q: %w", g.BPRAmount, err)
	}
	return s.q.CreateERAPaymentGroup(ctx, store.CreateERAPaymentGroupParams{
		GroupID:     g.GroupID,
		EraID:       g.EraID,
		LocationID:  g.LocationID,
		TraceNumber: g.TraceNumber,
		BprAmount:   bprNum,
		ClaimCount:  int32(g.ClaimCount),
		Claims:      g.Claims,
		Adjustments: g.Adjustments,
		Status:      g.Status,
	})
}

func (s *RealIngesterStore) SetERAPaymentGroupException(ctx context.Context, groupID string) error {
	return s.q.SetERAPaymentGroupException(ctx, groupID)
}

func (s *RealIngesterStore) MarkERAReconciliationTriggered(ctx context.Context, groupID string) error {
	return s.q.MarkERAReconciliationTriggered(ctx, groupID)
}

// ---- VCC payment groups ----

func (s *RealIngesterStore) GetActiveVCCPaymentGroup(ctx context.Context, locationID, traceID, fingerprint string) (*VCCPaymentGroup, error) {
	row, err := s.q.GetActiveVCCPaymentGroup(ctx, store.GetActiveVCCPaymentGroupParams{
		LocationID: locationID, TraceID: traceID, CardFingerprint: fingerprint,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return vccRowToGroup(row.GroupID, row.VccFileID, row.LocationID, row.TraceID, row.PaymentID,
		row.ProviderNpi.String, row.ProviderTaxID.String, row.CardFingerprint, row.TotalAmount,
		row.Status, row.IsAuthoritative, row.ReconciliationTriggeredAt, row.CreatedAt), nil
}

func (s *RealIngesterStore) GetMatchingVCCForERA(ctx context.Context, locationID, traceNumber string) (*VCCPaymentGroup, error) {
	row, err := s.q.GetMatchingVCCForERA(ctx, store.GetMatchingVCCForERAParams{
		LocationID: locationID, TraceID: traceNumber,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return vccRowToGroup(row.GroupID, row.VccFileID, row.LocationID, row.TraceID, row.PaymentID,
		row.ProviderNpi.String, row.ProviderTaxID.String, row.CardFingerprint, row.TotalAmount,
		row.Status, row.IsAuthoritative, row.ReconciliationTriggeredAt, row.CreatedAt), nil
}

func (s *RealIngesterStore) CreateVCCPaymentGroup(ctx context.Context, g *VCCPaymentGroup) error {
	totalNum, err := decimalStringToNumeric(g.TotalAmount)
	if err != nil {
		return fmt.Errorf("store CreateVCCPaymentGroup: invalid TotalAmount %q: %w", g.TotalAmount, err)
	}
	return s.q.CreateVCCPaymentGroup(ctx, store.CreateVCCPaymentGroupParams{
		GroupID:         g.GroupID,
		VccFileID:       g.VCCFileID,
		LocationID:      g.LocationID,
		TraceID:         g.TraceID,
		PaymentID:       g.PaymentID,
		ProviderNpi:     pgText(g.ProviderNPI),
		ProviderTaxID:   pgText(g.ProviderTaxID),
		CardFingerprint: g.CardFingerprint,
		TotalAmount:     totalNum,
		Status:          g.Status,
		IsAuthoritative: g.IsAuthoritative,
		PaymentMethodID: pgText(g.PaymentMethodID),
	})
}

func (s *RealIngesterStore) SetVCCPaymentGroupException(ctx context.Context, groupID string) error {
	return s.q.SetVCCPaymentGroupException(ctx, groupID)
}

func (s *RealIngesterStore) SupersedeVCCPaymentGroup(ctx context.Context, oldGroupID string, newGroup *VCCPaymentGroup) error {
	if err := s.q.MarkVCCGroupNonAuthoritative(ctx, oldGroupID); err != nil {
		return err
	}
	return s.CreateVCCPaymentGroup(ctx, newGroup)
}

func (s *RealIngesterStore) MarkVCCReconciliationTriggered(ctx context.Context, groupID string) error {
	return s.q.MarkVCCReconciliationTriggered(ctx, groupID)
}

// ---- VCC rows ----

func (s *RealIngesterStore) InsertVCCRow(ctx context.Context, r *VCCRow) error {
	issueDate, err := stringToDate(r.IssueDate)
	if err != nil {
		return fmt.Errorf("store InsertVCCRow: issue_date: %w", err)
	}
	amount, err := decimalStringToNumeric(r.Amount)
	if err != nil {
		return fmt.Errorf("store InsertVCCRow: amount: %w", err)
	}
	svcStart, err := stringToDate(r.ServiceDateStart)
	if err != nil {
		return fmt.Errorf("store InsertVCCRow: service_date_start: %w", err)
	}
	svcEnd, err := stringToDate(r.ServiceDateEnd)
	if err != nil {
		return fmt.Errorf("store InsertVCCRow: service_date_end: %w", err)
	}
	return s.q.InsertVCCRow(ctx, store.InsertVCCRowParams{
		RowID:             r.RowID,
		VccFileID:         r.VCCFileID,
		VccPaymentGroupID: pgText(r.VCCPaymentGroupID),
		LocationID:        r.LocationID,
		PaymentID:         r.PaymentID,
		TraceID:           r.TraceID,
		PayerName:         pgText(r.PayerName),
		ProviderNpi:       pgText(r.ProviderNPI),
		ProviderTaxID:     pgText(r.ProviderTaxID),
		IssueDate:         issueDate,
		Amount:            amount,
		CardFingerprint:   r.CardFingerprint,
		Last4:             r.Last4,
		ExpirationDate:    pgText(r.ExpirationDate),
		PatientID:         pgText(r.PatientID),
		ClaimID:           pgText(r.ClaimID),
		ServiceDateStart:  svcStart,
		ServiceDateEnd:    svcEnd,
	})
}

// ---- State transitions ----

func (s *RealIngesterStore) InsertTransition(ctx context.Context, p InsertTransitionParams) error {
	_, err := s.q.InsertStateTransition(ctx, store.InsertStateTransitionParams{
		TransitionID:   p.ID,
		EntityType:     p.EntityType,
		EntityID:       p.EntityID,
		FromState:      pgText(p.FromState),
		ToState:        p.ToState,
		TransitionedAt: pgNow(),
		Reason:         pgText(p.Reason),
	})
	return err
}

// ---- helpers ----

func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func pgNow() pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: time.Now(), Valid: true}
}

func decimalStringToNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if s == "" {
		return n, nil
	}
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("decimalStringToNumeric: %w", err)
	}
	return n, nil
}

func stringToDate(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("stringToDate: %w", err)
	}
	return pgtype.Date{Time: t, Valid: true, InfinityModifier: pgtype.Finite}, nil
}

func nullableTimestamp(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func eraRowToGroup(
	groupID, eraID, locationID, traceNumber, bprAmount string,
	claimCount int32,
	claims, adjustments []byte,
	status string,
	triggeredAt, createdAt pgtype.Timestamptz,
) *ERAPaymentGroup {
	return &ERAPaymentGroup{
		GroupID:                   groupID,
		EraID:                     eraID,
		LocationID:                locationID,
		TraceNumber:               traceNumber,
		BPRAmount:                 bprAmount,
		ClaimCount:                int(claimCount),
		Claims:                    append([]byte(nil), claims...),
		Adjustments:               append([]byte(nil), adjustments...),
		Status:                    status,
		ReconciliationTriggeredAt: nullableTimestamp(triggeredAt),
		CreatedAt:                 createdAt.Time,
	}
}

func vccRowToGroup(
	groupID, vccFileID, locationID, traceID, paymentID string,
	providerNPI, providerTaxID, cardFingerprint, totalAmount, status string,
	isAuthoritative bool,
	triggeredAt, createdAt pgtype.Timestamptz,
) *VCCPaymentGroup {
	return &VCCPaymentGroup{
		GroupID:                   groupID,
		VCCFileID:                 vccFileID,
		LocationID:                locationID,
		TraceID:                   traceID,
		PaymentID:                 paymentID,
		ProviderNPI:               providerNPI,
		ProviderTaxID:             providerTaxID,
		CardFingerprint:           cardFingerprint,
		TotalAmount:               totalAmount,
		PaymentMethodID:           "", // not selected by lookup queries; set at creation time
		Status:                    status,
		IsAuthoritative:           isAuthoritative,
		ReconciliationTriggeredAt: nullableTimestamp(triggeredAt),
		CreatedAt:                 createdAt.Time,
	}
}

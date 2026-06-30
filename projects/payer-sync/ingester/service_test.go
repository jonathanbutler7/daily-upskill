package ingester

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	seedersdk "github.com/jonathanbutler7/payer-sync-data-seeder/sdk"
)

// ─── Mock implementations ───────────────────────────────────────────────────

type mockClient struct {
	objects []seedersdk.ObjectMetadata
	bodies  map[string][]byte // key → encrypted bytes
	listErr error
	getErr  map[string]error
}

type savedRawFile struct {
	preferredKey string
	sourceKey    string
	contents     []byte
}

type mockRawStore struct {
	saves   []savedRawFile
	failFor map[string]error
}

func newMockRawStore() *mockRawStore {
	return &mockRawStore{failFor: make(map[string]error)}
}

func (m *mockRawStore) Save(_ context.Context, preferredKey, sourceKey string, encryptedBytes []byte) (string, error) {
	if err, ok := m.failFor[sourceKey]; ok {
		return "", err
	}
	cloned := append([]byte(nil), encryptedBytes...)
	m.saves = append(m.saves, savedRawFile{
		preferredKey: preferredKey,
		sourceKey:    sourceKey,
		contents:     cloned,
	})
	if preferredKey != "" {
		return preferredKey, nil
	}
	return "raw/" + sourceKey, nil
}

func (m *mockClient) ListObjects(_ context.Context, _ string) (seedersdk.ListObjectsResponse, error) {
	if m.listErr != nil {
		return seedersdk.ListObjectsResponse{}, m.listErr
	}
	return seedersdk.ListObjectsResponse{Objects: m.objects}, nil
}

func (m *mockClient) GetObject(_ context.Context, key string) (seedersdk.GetObjectResponse, error) {
	if m.getErr != nil {
		if err, ok := m.getErr[key]; ok {
			return seedersdk.GetObjectResponse{}, err
		}
	}
	body := m.bodies[key]
	// return metadata from the object list
	for _, obj := range m.objects {
		if obj.Key == key {
			return seedersdk.GetObjectResponse{Metadata: obj, EncryptedBytes: body}, nil
		}
	}
	return seedersdk.GetObjectResponse{EncryptedBytes: body}, nil
}

// noopDecryptor returns bytes as-is (test helper: the mock "encrypted" bytes ARE the plaintext)
type noopDecryptor struct{ failFor string }

func (d *noopDecryptor) Decrypt(b []byte, _, _ string) ([]byte, error) {
	if d.failFor != "" && strings.Contains(string(b), d.failFor) {
		return nil, errors.New("KMS unavailable")
	}
	return b, nil
}

// mockStore records every call so tests can assert on it.
type mockStore struct {
	eraHashes        map[string]bool // "locationID|hash" → exists
	vccHashes        map[string]bool
	eraGroups        map[string]*ERAPaymentGroup // "locationID|trace" → group
	vccGroups        map[string]*VCCPaymentGroup // "locationID|traceID|fingerprint" → group
	vccGroupsByTrace map[string]*VCCPaymentGroup // "locationID|trace" → group (for ERA→VCC matching)

	createdERAs      []CreateERAParams
	createdVCCFiles  []CreateVCCFileParams
	createdERAGroups []*ERAPaymentGroup
	createdVCCGroups []*VCCPaymentGroup
	vccRows          []*VCCRow
	transitions      []InsertTransitionParams
	eraExceptions    []string
	vccExceptions    []string
	supersessions    []string
	eraReconciled    []string
	vccReconciled    []string
}

func newMockStore() *mockStore {
	return &mockStore{
		eraHashes:        make(map[string]bool),
		vccHashes:        make(map[string]bool),
		eraGroups:        make(map[string]*ERAPaymentGroup),
		vccGroups:        make(map[string]*VCCPaymentGroup),
		vccGroupsByTrace: make(map[string]*VCCPaymentGroup),
	}
}

func (s *mockStore) ExistsERAByOfficeAndHash(_ context.Context, locationID, hash string) (bool, error) {
	return s.eraHashes[locationID+"|"+hash], nil
}

func (s *mockStore) CreateERA(_ context.Context, p CreateERAParams) error {
	s.createdERAs = append(s.createdERAs, p)
	return nil
}

func (s *mockStore) UpdateERAParsed(_ context.Context, p UpdateERAParsedParams) error {
	for i := range s.createdERAs {
		if s.createdERAs[i].ID == p.ID {
			s.createdERAs[i].PayerName = p.PayerName
			s.createdERAs[i].ProviderNPI = p.ProviderNPI
			s.createdERAs[i].ProviderTaxID = p.ProviderTaxID
			s.createdERAs[i].BPRAmount = p.BPRAmount
			s.createdERAs[i].PaymentMethod = p.PaymentMethod
			s.createdERAs[i].TraceNumber = p.TraceNumber
			s.createdERAs[i].Status = ERARemittanceStatusParsed
			return nil
		}
	}
	return fmt.Errorf("era %q not found", p.ID)
}

func (s *mockStore) SetERAParseFailure(_ context.Context, eraID string) error {
	for i := range s.createdERAs {
		if s.createdERAs[i].ID == eraID {
			s.createdERAs[i].Status = ERARemittanceStatusExceptionParseFailed
			return nil
		}
	}
	return fmt.Errorf("era %q not found", eraID)
}

func (s *mockStore) ExistsVCCByOfficeAndHash(_ context.Context, locationID, hash string) (bool, error) {
	return s.vccHashes[locationID+"|"+hash], nil
}

func (s *mockStore) CreateVCCFile(_ context.Context, p CreateVCCFileParams) error {
	s.createdVCCFiles = append(s.createdVCCFiles, p)
	return nil
}

func (s *mockStore) UpdateVCCFileParsed(_ context.Context, p UpdateVCCFileParsedParams) error {
	for i := range s.createdVCCFiles {
		if s.createdVCCFiles[i].ID == p.ID {
			s.createdVCCFiles[i].RowCount = p.RowCount
			s.createdVCCFiles[i].Status = VCCFileStatusParsed
			return nil
		}
	}
	return fmt.Errorf("vcc file %q not found", p.ID)
}

func (s *mockStore) SetVCCFileParseFailure(_ context.Context, vccFileID string, rowCount int) error {
	for i := range s.createdVCCFiles {
		if s.createdVCCFiles[i].ID == vccFileID {
			s.createdVCCFiles[i].RowCount = rowCount
			s.createdVCCFiles[i].Status = VCCFileStatusExceptionParseFailed
			return nil
		}
	}
	return fmt.Errorf("vcc file %q not found", vccFileID)
}

func (s *mockStore) GetActiveERAPaymentGroup(_ context.Context, locationID, trace string) (*ERAPaymentGroup, error) {
	return s.eraGroups[locationID+"|"+trace], nil
}

func (s *mockStore) CreateERAPaymentGroup(_ context.Context, g *ERAPaymentGroup) error {
	s.createdERAGroups = append(s.createdERAGroups, g)
	s.eraGroups[g.LocationID+"|"+g.TraceNumber] = g
	return nil
}

func (s *mockStore) SetERAPaymentGroupException(_ context.Context, id string) error {
	s.eraExceptions = append(s.eraExceptions, id)
	// update in-place so subsequent GetActive calls don't see it
	for k, g := range s.eraGroups {
		if g.GroupID == id {
			g.Status = ERAStatusException
			delete(s.eraGroups, k)
		}
	}
	return nil
}

func (s *mockStore) MarkERAReconciliationTriggered(_ context.Context, id string) error {
	s.eraReconciled = append(s.eraReconciled, id)
	return nil
}

func (s *mockStore) GetActiveVCCPaymentGroup(_ context.Context, locationID, traceID, fp string) (*VCCPaymentGroup, error) {
	return s.vccGroups[locationID+"|"+traceID+"|"+fp], nil
}

func (s *mockStore) GetMatchingVCCForERA(_ context.Context, locationID, trace string) (*VCCPaymentGroup, error) {
	return s.vccGroupsByTrace[locationID+"|"+trace], nil
}

func (s *mockStore) GetMatchingERAForVCC(_ context.Context, locationID, traceID string) (*ERAPaymentGroup, error) {
	return s.eraGroups[locationID+"|"+traceID], nil
}

func (s *mockStore) CreateVCCPaymentGroup(_ context.Context, g *VCCPaymentGroup) error {
	s.createdVCCGroups = append(s.createdVCCGroups, g)
	s.vccGroups[g.LocationID+"|"+g.TraceID+"|"+g.CardFingerprint] = g
	s.vccGroupsByTrace[g.LocationID+"|"+g.TraceID] = g
	return nil
}

func (s *mockStore) SetVCCPaymentGroupException(_ context.Context, id string) error {
	s.vccExceptions = append(s.vccExceptions, id)
	for k, g := range s.vccGroups {
		if g.GroupID == id {
			g.Status = VCCStatusException
			delete(s.vccGroups, k)
		}
	}
	return nil
}

func (s *mockStore) SupersedeVCCPaymentGroup(_ context.Context, oldID string, newGroup *VCCPaymentGroup) error {
	s.supersessions = append(s.supersessions, oldID)
	s.createdVCCGroups = append(s.createdVCCGroups, newGroup)
	s.vccGroups[newGroup.LocationID+"|"+newGroup.TraceID+"|"+newGroup.CardFingerprint] = newGroup
	s.vccGroupsByTrace[newGroup.LocationID+"|"+newGroup.TraceID] = newGroup
	return nil
}

func (s *mockStore) MarkVCCReconciliationTriggered(_ context.Context, id string) error {
	s.vccReconciled = append(s.vccReconciled, id)
	return nil
}

func (s *mockStore) InsertVCCRow(_ context.Context, r *VCCRow) error {
	s.vccRows = append(s.vccRows, r)
	return nil
}

func (s *mockStore) InsertTransition(_ context.Context, p InsertTransitionParams) error {
	s.transitions = append(s.transitions, p)
	return nil
}

type mockTrigger struct {
	calls []string // "eraID|vccID"
	err   error
}

func (t *mockTrigger) Trigger(_ context.Context, req ReconcileTriggerRequest) error {
	if t.err != nil {
		return t.err
	}
	t.calls = append(t.calls, req.ERAGroupID+"|"+req.VCCGroupID)
	return nil
}

type mockTokenizer struct {
	pmID  string
	err   error
	calls []tokenizerCall
}

type tokenizerCall struct {
	cardNumber string
	expMonth   string
	expYear    string
	cvv        string
}

func (t *mockTokenizer) CreatePaymentMethod(_ context.Context, cardNumber, expMonth, expYear, cvv string) (string, error) {
	t.calls = append(t.calls, tokenizerCall{
		cardNumber: cardNumber,
		expMonth:   expMonth,
		expYear:    expYear,
		cvv:        cvv,
	})
	if t.err != nil {
		return "", t.err
	}
	if t.pmID != "" {
		return t.pmID, nil
	}
	return "pm_test_mock", nil
}

func newMockTokenizer() *mockTokenizer {
	return &mockTokenizer{}
}

// ─── Test fixtures ──────────────────────────────────────────────────────────

const validERAContent = `ISA*00*          *00*          *ZZ*DELTADENTALCA   *ZZ*ACMEDENTAL      *260524*1200*^*00501*000000001*0*P*:~
BPR*I*450.00*C*VCC~
TRN*1*9876543210*9876543210~
N1*PR*DELTA DENTAL OF CALIFORNIA~
N1*PE*ACME DENTAL GROUP*XX*1234567890~
REF*TJ*12-3456789~
CLP*CLM-001*1*250.00*250.00~
CAS*CO*45*15.00*1~
SVC*HC:D0120*250.00*250.00~
CLP*CLM-002*1*200.00*200.00~
SVC*HC:D0140*200.00*200.00~
SE*9*0001~`

const validVCCContent = `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end
PMT-A,9876543210,DELTA DENTAL OF CALIFORNIA,1234567890,12-3456789,2026-05-24,250.00,4111111111111111,2028-01,123,John Doe,PAT-001,CLM-001,2026-04-01,2026-04-01
PMT-A,9876543210,DELTA DENTAL OF CALIFORNIA,1234567890,12-3456789,2026-05-24,200.00,4111111111111111,2028-01,123,Jane Smith,PAT-002,CLM-002,2026-04-02,2026-04-02`

const multiTraceVCCContent2 = `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end
PMT-A,9876543210,DELTA DENTAL OF CALIFORNIA,1234567890,12-3456789,2026-05-24,450.00,4111111111111111,2028-01,123,John Doe,PAT-001,CLM-001,2026-04-01,2026-04-01
PMT-B,1111111111,AETNA,1234567890,12-3456789,2026-05-24,300.00,4222222222222222,2029-06,456,Jane Smith,PAT-002,CLM-002,2026-04-15,2026-04-15`

func newERAObject(key, hash string) seedersdk.ObjectMetadata {
	return seedersdk.ObjectMetadata{
		Key: key, FileType: FileTypeERA, SourceFilename: key,
		ReceivedAt: time.Now(), RawStorageKey: "raw/" + key,
		Encryption: seedersdk.EncryptionDetails{PlaintextSHA256: hash},
	}
}

func newVCCObject(key, hash string) seedersdk.ObjectMetadata {
	return seedersdk.ObjectMetadata{
		Key: key, FileType: FileTypeVCC, SourceFilename: key,
		ReceivedAt: time.Now(), RawStorageKey: "raw/" + key,
		Encryption: seedersdk.EncryptionDetails{PlaintextSHA256: hash},
	}
}

func newSvc(client *mockClient, store *mockStore, trigger *mockTrigger) *Service {
	return NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, newMockRawStore(), store, trigger, newMockTokenizer(),
	)
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// ING-001: unknown file type is recorded as an error but ERA/VCC still ingest.
func TestService_ING001_UnknownFileTypeDoesNotBlockOthers(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{
			newERAObject("era-good.edi", "hash-era"),
			{Key: "mystery.xml", FileType: "UNKNOWN", Encryption: seedersdk.EncryptionDetails{PlaintextSHA256: "hash-unk"}},
			newVCCObject("vcc-good.csv", "hash-vcc"),
		},
		bodies: map[string][]byte{
			"era-good.edi": []byte(validERAContent),
			"vcc-good.csv": []byte(validVCCContent),
		},
	}
	svc := newSvc(client, st, trig)
	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
	// ERA and VCC both processed
	if result.Processed != 2 {
		t.Errorf("Processed = %d, want 2", result.Processed)
	}
	// Unknown file is an error, not a silent skip
	if len(result.Errors) != 1 {
		t.Errorf("Errors count = %d, want 1", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error(), "mystery.xml") {
		t.Errorf("error should reference mystery.xml: %v", result.Errors[0])
	}
}

// ING-003 / ING-006: pre-decrypt dedup — file with known plaintext hash is skipped before GetObject.
func TestService_ING003_PreDecryptDedupSkipsGetObject(t *testing.T) {
	st := newMockStore()
	st.eraHashes["office-001|hash-era"] = true // already ingested

	getObjectCalled := false
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era-dup.edi", "hash-era")},
		bodies:  map[string][]byte{},
	}
	// Override GetObject to detect if it's called
	customClient := &trackingClient{inner: client, onGet: func(key string) { getObjectCalled = true }}

	trig := &mockTrigger{}
	svc := NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		customClient, &noopDecryptor{}, newMockRawStore(), st, trig, newMockTokenizer(),
	)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected error: %v %v", err, result.Errors)
	}
	if getObjectCalled {
		t.Error("GetObject should NOT have been called for a pre-decrypt duplicate")
	}
	if len(st.createdERAs) != 0 {
		t.Error("no ERA record should be created for a duplicate")
	}
	if len(trig.calls) != 0 {
		t.Error("no reconciliation trigger should fire for a duplicate")
	}
}

// ING-004: ERA happy path — parses, persists, creates payment group, triggers reconciliation.
func TestService_ING004_ERAHappyPath(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	rawStore := newMockRawStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era-9876543210.edi", "hash-era-001")},
		bodies:  map[string][]byte{"era-9876543210.edi": []byte(validERAContent)},
	}
	svc := NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, rawStore, st, trig, newMockTokenizer(),
	)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v %v", err, result.Errors)
	}

	// Normalized ERA persisted
	if len(st.createdERAs) != 1 {
		t.Fatalf("createdERAs count = %d, want 1", len(st.createdERAs))
	}
	era := st.createdERAs[0]
	if era.TraceNumber != "9876543210" {
		t.Errorf("TraceNumber = %q, want 9876543210", era.TraceNumber)
	}
	if era.PayerName != "DELTA DENTAL OF CALIFORNIA" {
		t.Errorf("PayerName = %q, want DELTA DENTAL OF CALIFORNIA", era.PayerName)
	}
	if era.BPRAmount != "450.00" {
		t.Errorf("BPRAmount = %q, want 450.00", era.BPRAmount)
	}
	if era.LocationID != "office-001" {
		t.Errorf("LocationID = %q, want office-001", era.LocationID)
	}
	if era.FileHash != "hash-era-001" {
		t.Errorf("FileHash = %q, want hash-era-001", era.FileHash)
	}
	if era.StorageKey != "raw/era-9876543210.edi" {
		t.Errorf("StorageKey = %q, want raw/era-9876543210.edi", era.StorageKey)
	}
	if era.Status != ERARemittanceStatusParsed {
		t.Errorf("Status = %q, want %s", era.Status, ERARemittanceStatusParsed)
	}
	if len(rawStore.saves) != 1 {
		t.Fatalf("raw store saves = %d, want 1", len(rawStore.saves))
	}
	if got := string(rawStore.saves[0].contents); got != validERAContent {
		t.Errorf("raw stored content mismatch: got %q", got)
	}

	// Payment group created in AWAITING_VCC
	if len(st.createdERAGroups) != 1 {
		t.Fatalf("createdERAGroups count = %d, want 1", len(st.createdERAGroups))
	}
	if st.createdERAGroups[0].Status != ERAStatusAwaitingVCC {
		t.Errorf("group status = %q, want AWAITING_VCC", st.createdERAGroups[0].Status)
	}
	if st.createdERAGroups[0].ClaimCount != 2 {
		t.Errorf("claim count = %d, want 2", st.createdERAGroups[0].ClaimCount)
	}
	if !strings.Contains(string(st.createdERAGroups[0].Claims), "CLM-001") {
		t.Errorf("claims json = %s, want CLM-001", string(st.createdERAGroups[0].Claims))
	}
	if !strings.Contains(string(st.createdERAGroups[0].Adjustments), "\"ReasonCode\":\"45\"") {
		t.Errorf("adjustments json = %s, want reason code 45", string(st.createdERAGroups[0].Adjustments))
	}

	// Reconciliation trigger emitted
	if len(trig.calls) != 1 {
		t.Errorf("trigger calls = %d, want 1", len(trig.calls))
	}

	// Audit trail has at least one transition
	if len(st.transitions) == 0 {
		t.Error("expected at least one state transition in audit trail")
	}
	var sawRaw, sawParsed bool
	for _, tr := range st.transitions {
		if tr.EntityType != EntityTypeERARemittance {
			continue
		}
		if tr.ToState == ERARemittanceStatusReceivedRaw {
			sawRaw = true
		}
		if tr.ToState == ERARemittanceStatusParsed {
			sawParsed = true
		}
	}
	if !sawRaw || !sawParsed {
		t.Errorf("expected ERA remittance transitions to include RECEIVED_RAW and PARSED, got %+v", st.transitions)
	}
}

// ING-005: VCC happy path — two rows grouped, amount summed, raw PAN absent.
func TestService_ING005_VCCHappyPath(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc-9876543210.csv", "hash-vcc-001")},
		bodies:  map[string][]byte{"vcc-9876543210.csv": []byte(validVCCContent)},
	}
	svc := newSvc(client, st, trig)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v %v", err, result.Errors)
	}

	// VCC file persisted
	if len(st.createdVCCFiles) != 1 {
		t.Fatalf("createdVCCFiles count = %d, want 1", len(st.createdVCCFiles))
	}
	if st.createdVCCFiles[0].RowCount != 2 {
		t.Errorf("RowCount = %d, want 2", st.createdVCCFiles[0].RowCount)
	}
	if st.createdVCCFiles[0].Status != VCCFileStatusParsed {
		t.Errorf("Status = %q, want %s", st.createdVCCFiles[0].Status, VCCFileStatusParsed)
	}

	// One payment group
	if len(st.createdVCCGroups) != 1 {
		t.Fatalf("createdVCCGroups count = %d, want 1", len(st.createdVCCGroups))
	}
	if st.createdVCCGroups[0].TotalAmount != "450.00" {
		t.Errorf("TotalAmount = %q, want 450.00", st.createdVCCGroups[0].TotalAmount)
	}

	// Two VCC rows, neither contains raw PAN
	if len(st.vccRows) != 2 {
		t.Fatalf("vccRows count = %d, want 2", len(st.vccRows))
	}
	for i, row := range st.vccRows {
		if row.CardFingerprint == "4111111111111111" {
			t.Errorf("row %d: CardFingerprint is raw PAN — ING-013 violation", i)
		}
		if !strings.HasPrefix(row.CardFingerprint, "fp:") {
			t.Errorf("row %d: CardFingerprint %q not in fp: format", i, row.CardFingerprint)
		}
		if row.Last4 != "1111" {
			t.Errorf("row %d: Last4 = %q, want 1111", i, row.Last4)
		}
	}

	// Reconciliation trigger emitted
	if len(trig.calls) != 1 {
		t.Errorf("trigger calls = %d, want 1", len(trig.calls))
	}
}

// ING-007: parse failure — no ERA payment group created, no trigger, audit entry recorded.
func TestService_ING007_ERAParseFailure(t *testing.T) {
	corruptERA := `BPR*I*NOTANUMBER*C*VCC~
TRN*1*9876543210~
SE*2*0001~`
	st := newMockStore()
	trig := &mockTrigger{}
	rawStore := newMockRawStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era-corrupt.edi", "hash-corrupt")},
		bodies:  map[string][]byte{"era-corrupt.edi": []byte(corruptERA)},
	}
	svc := NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, rawStore, st, trig, newMockTokenizer(),
	)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("parse failure should be non-fatal: %v %v", err, result.Errors)
	}

	// ERA record stored (for raw replay), but as exception
	if len(st.createdERAs) != 1 {
		t.Fatalf("expected ERA record for failed parse, got %d", len(st.createdERAs))
	}
	if st.createdERAs[0].Status != ERARemittanceStatusExceptionParseFailed {
		t.Errorf("status = %q, want %s", st.createdERAs[0].Status, ERARemittanceStatusExceptionParseFailed)
	}
	if st.createdERAs[0].StorageKey != "raw/era-corrupt.edi" {
		t.Errorf("storage key = %q, want raw/era-corrupt.edi", st.createdERAs[0].StorageKey)
	}
	if len(rawStore.saves) != 1 {
		t.Fatalf("raw store saves = %d, want 1", len(rawStore.saves))
	}
	// No payment group
	if len(st.createdERAGroups) != 0 {
		t.Errorf("no payment group should be created on parse failure, got %d", len(st.createdERAGroups))
	}
	// No trigger
	if len(trig.calls) != 0 {
		t.Error("no reconciliation trigger should fire on parse failure")
	}
	// Audit transition with reason
	var hasFailureTransition bool
	for _, tr := range st.transitions {
		if tr.ToState == ERARemittanceStatusExceptionParseFailed && strings.Contains(tr.Reason, "BPR") {
			hasFailureTransition = true
		}
	}
	if !hasFailureTransition {
		t.Error("expected audit transition recording parse failure reason")
	}
}

// ING-009: exact-duplicate VCC file (same bytes = same hash) is silently skipped.
func TestService_ING009_ExactDuplicateVCCSkipped(t *testing.T) {
	st := newMockStore()
	st.vccHashes["office-001|hash-vcc-001"] = true // already ingested
	trig := &mockTrigger{}
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc-dup.csv", "hash-vcc-001")},
		bodies:  map[string][]byte{},
	}
	svc := newSvc(client, st, trig)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v %v", err, result.Errors)
	}
	if len(st.createdVCCFiles) != 0 {
		t.Error("no VCC file record should be created for an exact duplicate")
	}
	if len(trig.calls) != 0 {
		t.Error("no trigger should fire for an exact duplicate")
	}
}

// ING-010: safe VCC correction (same fingerprint+trace, different non-funding metadata, group AWAITING_ERA) → supersede.
func TestService_ING010_SafeVCCCorrectionSupersedes(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}

	// Seed existing group: same card, same amount, AWAITING_ERA
	fp := CardFingerprint("4111111111111111", "test-key")
	existingGroup := &VCCPaymentGroup{
		GroupID: "existing-grp", VCCFileID: "old-file", LocationID: "office-001",
		TraceID: "9876543210", PaymentID: "PMT-A", CardFingerprint: fp,
		TotalAmount: "450.00", Status: VCCStatusAwaitingERA, IsAuthoritative: true,
	}
	st.vccGroups["office-001|9876543210|"+fp] = existingGroup
	st.vccGroupsByTrace["office-001|9876543210"] = existingGroup

	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc-corrected.csv", "hash-corrected")},
		bodies:  map[string][]byte{"vcc-corrected.csv": []byte(validVCCContent)},
	}
	svc := newSvc(client, st, trig)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v %v", err, result.Errors)
	}

	// Old group was superseded
	if len(st.supersessions) != 1 || st.supersessions[0] != "existing-grp" {
		t.Errorf("expected supersession of existing-grp, got %v", st.supersessions)
	}
	if len(st.vccRows) != 2 {
		t.Errorf("expected corrected rows to be persisted, got %d", len(st.vccRows))
	}
	if len(trig.calls) != 1 {
		t.Errorf("expected reconciler trigger for superseded authoritative group, got %d calls", len(trig.calls))
	}
	// No exception
	if len(st.vccExceptions) != 0 {
		t.Errorf("no exception expected for safe correction, got %v", st.vccExceptions)
	}
}

// ING-011: material VCC conflict (different amount) → exception, no trigger.
func TestService_ING011_MaterialVCCConflictException(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	fp := CardFingerprint("4111111111111111", "test-key")

	// Existing group with 450.00
	existingGroup := &VCCPaymentGroup{
		GroupID: "existing-grp", VCCFileID: "old-file", LocationID: "office-001",
		TraceID: "9876543210", PaymentID: "PMT-A", CardFingerprint: fp,
		TotalAmount: "500.00", // different! conflict
		Status:      VCCStatusAwaitingERA, IsAuthoritative: true,
	}
	st.vccGroups["office-001|9876543210|"+fp] = existingGroup

	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc-amended.csv", "hash-amended")},
		bodies:  map[string][]byte{"vcc-amended.csv": []byte(validVCCContent)}, // new file has 450.00
	}
	svc := newSvc(client, st, trig)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v %v", err, result.Errors)
	}

	// New group should be created as EXCEPTION
	var exceptionGroup *VCCPaymentGroup
	for _, g := range st.createdVCCGroups {
		if g.Status == VCCStatusException {
			exceptionGroup = g
		}
	}
	if exceptionGroup == nil {
		t.Fatal("expected a VCC payment group in EXCEPTION status")
	}
	// No trigger emitted
	if len(trig.calls) != 0 {
		t.Error("no reconciliation trigger should fire on material conflict")
	}
}

// ING-012a: ERA arrives first, VCC not yet present → trigger fires with ERA group only.
func TestService_ING012a_ERAArrivesFirst_TriggerFires(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	// No VCC present in store yet

	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era.edi", "hash-era-new")},
		bodies:  map[string][]byte{"era.edi": []byte(validERAContent)},
	}
	svc := newSvc(client, st, trig)
	svc.Run(context.Background())

	// ERA payment group created in AWAITING_VCC
	if len(st.createdERAGroups) != 1 || st.createdERAGroups[0].Status != ERAStatusAwaitingVCC {
		t.Errorf("expected one AWAITING_VCC ERA group, got %+v", st.createdERAGroups)
	}
	// Trigger still fires so reconciler can attempt matching
	if len(trig.calls) != 1 {
		t.Errorf("expected 1 trigger call, got %d", len(trig.calls))
	}
}

// ING-012b: VCC arrives first, ERA not yet present → trigger fires with VCC group only.
func TestService_ING012b_VCCArrivesFirst_TriggerFires(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	// No ERA present in store yet

	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc.csv", "hash-vcc-new")},
		bodies:  map[string][]byte{"vcc.csv": []byte(validVCCContent)},
	}
	svc := newSvc(client, st, trig)
	svc.Run(context.Background())

	if len(st.createdVCCGroups) != 1 || st.createdVCCGroups[0].Status != VCCStatusAwaitingERA {
		t.Errorf("expected one AWAITING_ERA VCC group, got %+v", st.createdVCCGroups)
	}
	if len(trig.calls) != 1 {
		t.Errorf("expected 1 trigger call, got %d", len(trig.calls))
	}
}

// ING-012c: duplicate ingest does NOT trigger reconciliation.
func TestService_ING012c_DuplicateIngestNoTrigger(t *testing.T) {
	st := newMockStore()
	st.eraHashes["office-001|hash-era-dup"] = true // already ingested
	trig := &mockTrigger{}
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era-dup.edi", "hash-era-dup")},
	}
	svc := newSvc(client, st, trig)
	svc.Run(context.Background())

	if len(trig.calls) != 0 {
		t.Errorf("no trigger should fire for duplicate, got %d calls", len(trig.calls))
	}
}

// ING-016: VCC file with two trace IDs produces two independent payment groups.
func TestService_ING016_MultiTraceVCCProducesIndependentGroups(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc-multi.csv", "hash-multi")},
		bodies:  map[string][]byte{"vcc-multi.csv": []byte(multiTraceVCCContent2)},
	}
	svc := newSvc(client, st, trig)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v %v", err, result.Errors)
	}

	if len(st.createdVCCGroups) != 2 {
		t.Fatalf("expected 2 VCC payment groups, got %d", len(st.createdVCCGroups))
	}
	amounts := map[string]string{}
	for _, g := range st.createdVCCGroups {
		amounts[g.TraceID] = g.TotalAmount
	}
	if amounts["9876543210"] != "450.00" {
		t.Errorf("trace 9876543210 amount = %q, want 450.00", amounts["9876543210"])
	}
	if amounts["1111111111"] != "300.00" {
		t.Errorf("trace 1111111111 amount = %q, want 300.00", amounts["1111111111"])
	}
	// Two independent triggers
	if len(trig.calls) != 2 {
		t.Errorf("expected 2 reconciliation triggers (one per group), got %d", len(trig.calls))
	}
}

// ING-017: two distinct ERAs with the same trace number → both become EXCEPTION.
func TestService_ING017_DuplicateTraceERAException(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}

	// Seed an existing active ERA payment group for the same trace
	existing := &ERAPaymentGroup{
		GroupID: "era-grp-existing", EraID: "era-existing", LocationID: "office-001",
		TraceNumber: "9876543210", BPRAmount: "450.00", Status: ERAStatusAwaitingVCC,
	}
	st.eraGroups["office-001|9876543210"] = existing

	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era-new.edi", "hash-era-new")},
		bodies:  map[string][]byte{"era-new.edi": []byte(validERAContent)},
	}
	svc := newSvc(client, st, trig)
	result, err := svc.Run(context.Background())
	if err != nil || len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v %v", err, result.Errors)
	}

	// Original group should be moved to EXCEPTION
	if len(st.eraExceptions) == 0 {
		t.Error("expected original ERA payment group to be set to EXCEPTION")
	}
	var foundOriginal bool
	for _, id := range st.eraExceptions {
		if id == "era-grp-existing" {
			foundOriginal = true
		}
	}
	if !foundOriginal {
		t.Error("expected era-grp-existing to be set to EXCEPTION")
	}
	// New group also created as EXCEPTION
	var hasNewException bool
	for _, g := range st.createdERAGroups {
		if g.Status == ERAStatusException && g.TraceNumber == "9876543210" {
			hasNewException = true
		}
	}
	if !hasNewException {
		t.Error("expected new ERA group for duplicate trace to be created as EXCEPTION")
	}
	// No reconciliation trigger
	if len(trig.calls) != 0 {
		t.Errorf("no trigger should fire when trace conflict is detected, got %d calls", len(trig.calls))
	}
}

// ING-018: remote server unavailable → Run returns error immediately (list fails).
func TestService_ING018_RemoteServerUnavailable(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	client := &mockClient{listErr: errors.New("connection refused")}
	svc := newSvc(client, st, trig)
	_, err := svc.Run(context.Background())
	if err == nil {
		t.Fatal("expected error when remote server is unavailable")
	}
	if !strings.Contains(err.Error(), "list objects") {
		t.Errorf("error should mention list objects: %v", err)
	}
}

// ING-019: KMS/decryptor fails → file is an error, others continue.
func TestService_ING019_DecryptorFails_OtherFilesContinue(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	rawStore := newMockRawStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{
			newERAObject("era-bad.edi", "hash-bad"),
			newVCCObject("vcc-good.csv", "hash-vcc"),
		},
		bodies: map[string][]byte{
			"era-bad.edi":  []byte("SENTINEL_FAIL"), // noopDecryptor will fail for this
			"vcc-good.csv": []byte(validVCCContent),
		},
	}
	svc := NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		client,
		&noopDecryptor{failFor: "SENTINEL_FAIL"},
		rawStore,
		st, trig, newMockTokenizer(),
	)
	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected fatal error: %v", err)
	}
	// ERA download errored, VCC succeeded
	if len(result.Errors) != 1 {
		t.Errorf("Errors count = %d, want 1", len(result.Errors))
	}
	if result.Processed != 1 {
		t.Errorf("Processed = %d, want 1", result.Processed)
	}
	if len(rawStore.saves) != 2 {
		t.Errorf("raw store saves = %d, want 2", len(rawStore.saves))
	}
}

func TestService_RawStoreFailure_OtherFilesContinue(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	rawStore := newMockRawStore()
	rawStore.failFor["era-bad.edi"] = errors.New("disk full")
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{
			newERAObject("era-bad.edi", "hash-era-bad"),
			newVCCObject("vcc-good.csv", "hash-vcc-good"),
		},
		bodies: map[string][]byte{
			"era-bad.edi":  []byte(validERAContent),
			"vcc-good.csv": []byte(validVCCContent),
		},
	}
	svc := NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, rawStore, st, trig, newMockTokenizer(),
	)
	result, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected fatal error: %v", err)
	}
	if result.Processed != 1 {
		t.Errorf("Processed = %d, want 1", result.Processed)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("Errors count = %d, want 1", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error(), "disk full") {
		t.Errorf("error = %v, want disk full", result.Errors[0])
	}
	if len(st.createdERAs) != 0 {
		t.Errorf("createdERAs = %d, want 0", len(st.createdERAs))
	}
	if len(st.createdVCCFiles) != 1 {
		t.Errorf("createdVCCFiles = %d, want 1", len(st.createdVCCFiles))
	}
}

// ING-020: location isolation — same trace for two different locations produces two independent groups.
func TestService_ING020_OfficeIsolation(t *testing.T) {
	// Location-001 service
	st1 := newMockStore()
	trig1 := &mockTrigger{}
	client1 := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era.edi", "hash-era")},
		bodies:  map[string][]byte{"era.edi": []byte(validERAContent)},
	}
	svc1 := NewService(Config{LocationID: "office-001", FingerprintKey: "test-key"}, client1, &noopDecryptor{}, newMockRawStore(), st1, trig1, newMockTokenizer())
	svc1.Run(context.Background())

	// Location-002 service (completely separate store)
	st2 := newMockStore()
	trig2 := &mockTrigger{}
	client2 := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era.edi", "hash-era")},
		bodies:  map[string][]byte{"era.edi": []byte(validERAContent)},
	}
	svc2 := NewService(Config{LocationID: "office-002", FingerprintKey: "test-key"}, client2, &noopDecryptor{}, newMockRawStore(), st2, trig2, newMockTokenizer())
	svc2.Run(context.Background())

	// Each location gets its own group
	if len(st1.createdERAGroups) != 1 || st1.createdERAGroups[0].LocationID != "office-001" {
		t.Errorf("office-001 should have its own ERA group: %+v", st1.createdERAGroups)
	}
	if len(st2.createdERAGroups) != 1 || st2.createdERAGroups[0].LocationID != "office-002" {
		t.Errorf("office-002 should have its own ERA group: %+v", st2.createdERAGroups)
	}
	// office-001 hash not treated as dup in office-002's store
	if st2.eraHashes["office-002|hash-era"] {
		t.Error("office-002 should not see office-001's records as duplicates")
	}
}

// ING-023: reconciliation trigger fires exactly once even on crash/restart simulation.
func TestService_ING023_ReconciliationTriggerExactlyOnce(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	fp := CardFingerprint("4111111111111111", "test-key")

	// Simulate: ERA group already exists and was already matched but trigger wasn't stamped yet
	now := time.Now()
	existingERAGroup := &ERAPaymentGroup{
		GroupID: "era-grp-1", EraID: "era-1", LocationID: "office-001",
		TraceNumber: "9876543210", BPRAmount: "450.00", Status: ERAStatusMatched,
		ReconciliationTriggeredAt: &now, // already triggered
	}
	st.eraGroups["office-001|9876543210"] = existingERAGroup

	// VCC group with matching trace arrives
	existingVCCGroup := &VCCPaymentGroup{
		GroupID: "vcc-grp-1", VCCFileID: "vcc-1", LocationID: "office-001",
		TraceID: "9876543210", CardFingerprint: fp,
		TotalAmount: "450.00", Status: VCCStatusAwaitingERA, IsAuthoritative: true,
		ReconciliationTriggeredAt: nil, // NOT yet stamped
	}
	st.vccGroupsByTrace["office-001|9876543210"] = existingVCCGroup

	// Now ingest the same ERA again (simulating restart)
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era.edi", "hash-era-new")},
		bodies:  map[string][]byte{"era.edi": []byte(validERAContent)},
	}
	svc := newSvc(client, st, trig)
	svc.Run(context.Background())

	// ERA group already has reconciliation_triggered_at set — trigger must NOT fire again
	if len(trig.calls) != 0 {
		t.Errorf("trigger should not fire again when ERA group already has reconciliation_triggered_at, got %d calls", len(trig.calls))
	}
}

// ING-013: confirms PAN and CVV do not appear in any persisted field (already in parser tests,
// but this integration-level test verifies the whole pipeline path through the service).
func TestService_ING013_PANNotPersistedByService(t *testing.T) {
	st := newMockStore()
	trig := &mockTrigger{}
	tok := newMockTokenizer()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc.csv", "hash-vcc")},
		bodies:  map[string][]byte{"vcc.csv": []byte(validVCCContent)},
	}
	svc := NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, newMockRawStore(), st, trig, tok,
	)
	svc.Run(context.Background())

	if len(tok.calls) != 1 {
		t.Fatalf("tokenizer calls = %d, want 1", len(tok.calls))
	}
	if got := tok.calls[0].cvv; got != "" {
		t.Fatalf("tokenizer cvv = %q, want empty", got)
	}
	if got := tok.calls[0].cardNumber; got == "" {
		t.Fatal("tokenizer card number was empty")
	}

	if len(st.createdVCCGroups) == 0 {
		t.Fatal("expected at least one VCC payment group")
	}
	if got := st.createdVCCGroups[0].PaymentMethodID; got == "" {
		t.Fatal("expected PaymentMethodID to be persisted on VCC payment group")
	}

	for i, row := range st.vccRows {
		if strings.Contains(row.CardFingerprint, "4111111111111111") {
			t.Errorf("row %d: raw PAN in CardFingerprint", i)
		}
		// last4 is acceptable
		if row.Last4 != "1111" {
			t.Errorf("row %d: Last4 = %q, want 1111", i, row.Last4)
		}
	}
}

// ─── ING-024: NPI validation ────────────────────────────────────────────────

// ING-024a: ERA with matching NPI passes when ExpectedNPI is set.
func TestService_ING024a_ERAMatchingNPIPasses(t *testing.T) {
	st := newMockStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era.edi", "hash-era")},
		bodies:  map[string][]byte{"era.edi": []byte(validERAContent)},
	}
	svc := NewService(
		Config{LocationID: "office-001", ExpectedNPI: "1234567890", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, newMockRawStore(), st, &mockTrigger{}, newMockTokenizer(),
	)
	result, _ := svc.Run(context.Background())
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %s", formatErrors(result.Errors))
	}
	if len(st.createdERAs) != 1 {
		t.Errorf("expected 1 ERA created, got %d", len(st.createdERAs))
	}
}

// ING-024b: ERA with mismatched NPI is rejected when ExpectedNPI is set.
func TestService_ING024b_ERAMismatchedNPIRejected(t *testing.T) {
	st := newMockStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era.edi", "hash-era")},
		bodies:  map[string][]byte{"era.edi": []byte(validERAContent)}, // NPI is 1234567890
	}
	svc := NewService(
		Config{LocationID: "office-001", ExpectedNPI: "9999999999", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, newMockRawStore(), st, &mockTrigger{}, newMockTokenizer(),
	)
	result, _ := svc.Run(context.Background())
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error for NPI mismatch, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Reason.Error(), "NPI mismatch") {
		t.Errorf("expected NPI mismatch error, got: %v", result.Errors[0].Reason)
	}
	if len(st.createdERAs) != 1 {
		t.Errorf("expected raw ERA record to be retained on NPI mismatch, got %d", len(st.createdERAs))
	}
	if len(st.createdERAs) == 1 && st.createdERAs[0].Status != ERARemittanceStatusReceivedRaw {
		t.Errorf("ERA status = %q, want %s after NPI mismatch", st.createdERAs[0].Status, ERARemittanceStatusReceivedRaw)
	}
}

// ING-024c: VCC with matching NPI passes when ExpectedNPI is set.
func TestService_ING024c_VCCMatchingNPIPasses(t *testing.T) {
	st := newMockStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc.csv", "hash-vcc")},
		bodies:  map[string][]byte{"vcc.csv": []byte(validVCCContent)}, // NPI is 1234567890
	}
	svc := NewService(
		Config{LocationID: "office-001", ExpectedNPI: "1234567890", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, newMockRawStore(), st, &mockTrigger{}, newMockTokenizer(),
	)
	result, _ := svc.Run(context.Background())
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %s", formatErrors(result.Errors))
	}
}

// ING-024d: VCC with mismatched NPI is rejected when ExpectedNPI is set.
func TestService_ING024d_VCCMismatchedNPIRejected(t *testing.T) {
	st := newMockStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newVCCObject("vcc.csv", "hash-vcc")},
		bodies:  map[string][]byte{"vcc.csv": []byte(validVCCContent)}, // NPI is 1234567890
	}
	svc := NewService(
		Config{LocationID: "office-001", ExpectedNPI: "9999999999", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, newMockRawStore(), st, &mockTrigger{}, newMockTokenizer(),
	)
	result, _ := svc.Run(context.Background())
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error for NPI mismatch, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Reason.Error(), "NPI mismatch") {
		t.Errorf("expected NPI mismatch error, got: %v", result.Errors[0].Reason)
	}
	if len(st.createdVCCFiles) != 1 {
		t.Errorf("expected raw VCC record to be retained on NPI mismatch, got %d", len(st.createdVCCFiles))
	}
	if len(st.createdVCCFiles) == 1 && st.createdVCCFiles[0].Status != VCCFileStatusReceivedRaw {
		t.Errorf("VCC status = %q, want %s after NPI mismatch", st.createdVCCFiles[0].Status, VCCFileStatusReceivedRaw)
	}
}

// ING-024e: empty ExpectedNPI skips NPI validation entirely.
func TestService_ING024e_EmptyExpectedNPISkipsValidation(t *testing.T) {
	st := newMockStore()
	client := &mockClient{
		objects: []seedersdk.ObjectMetadata{newERAObject("era.edi", "hash-era")},
		bodies:  map[string][]byte{"era.edi": []byte(validERAContent)},
	}
	// No ExpectedNPI set — should pass regardless of the NPI in the file
	svc := NewService(
		Config{LocationID: "office-001", FingerprintKey: "test-key"},
		client, &noopDecryptor{}, newMockRawStore(), st, &mockTrigger{}, newMockTokenizer(),
	)
	result, _ := svc.Run(context.Background())
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors when ExpectedNPI is empty, got: %s", formatErrors(result.Errors))
	}
}

// ─── trackingClient helper ──────────────────────────────────────────────────
type trackingClient struct {
	inner *mockClient
	onGet func(string)
}

func (c *trackingClient) ListObjects(ctx context.Context, ft string) (seedersdk.ListObjectsResponse, error) {
	return c.inner.ListObjects(ctx, ft)
}

func (c *trackingClient) GetObject(ctx context.Context, key string) (seedersdk.GetObjectResponse, error) {
	if c.onGet != nil {
		c.onGet(key)
	}
	return c.inner.GetObject(ctx, key)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func formatErrors(errs []*FileError) string {
	if len(errs) == 0 {
		return "<none>"
	}
	var sb strings.Builder
	for _, e := range errs {
		fmt.Fprintf(&sb, "  %v\n", e)
	}
	return sb.String()
}

package ingester

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	seedersdk "github.com/jonathanbutler7/payer-sync-data-seeder/sdk"
)

// ---- Status constants ----

const (
	ERARemittanceStatusReceivedRaw          = "RECEIVED_RAW"
	ERARemittanceStatusParsed               = "PARSED"
	ERARemittanceStatusExceptionParseFailed = "EXCEPTION_PARSE_FAILED"

	VCCFileStatusReceivedRaw          = "RECEIVED_RAW"
	VCCFileStatusParsed               = "PARSED"
	VCCFileStatusExceptionParseFailed = "EXCEPTION_PARSE_FAILED"

	ERAStatusAwaitingVCC = "AWAITING_VCC"
	ERAStatusMatched     = "MATCHED"
	ERAStatusException   = "EXCEPTION"

	VCCStatusAwaitingERA = "AWAITING_ERA"
	VCCStatusMatched     = "MATCHED"
	VCCStatusException   = "EXCEPTION"

	EntityTypeERARemittance = "era_remittance"
	EntityTypeVCCFile       = "vcc_file"
	EntityTypeERAPaymentGrp = "era_payment_group"
	EntityTypeVCCPaymentGrp = "vcc_payment_group"

	FileTypeERA = "ERA"
	FileTypeVCC = "VCC"

	// allFileTypesFilter requests all object types from the remote API.
	allFileTypesFilter = ""
)

// ---- Domain types (not sqlc-generated; used for new tables) ----

type ERAPaymentGroup struct {
	GroupID                   string
	EraID                     string
	LocationID                string
	TraceNumber               string
	BPRAmount                 string
	ClaimCount                int
	Claims                    []byte
	Adjustments               []byte
	Status                    string
	ReconciliationTriggeredAt *time.Time
	CreatedAt                 time.Time
}

type VCCPaymentGroup struct {
	GroupID                   string
	VCCFileID                 string
	LocationID                string
	TraceID                   string
	PaymentID                 string
	ProviderNPI               string
	ProviderTaxID             string
	CardFingerprint           string
	TotalAmount               string
	PaymentMethodID           string
	Status                    string
	IsAuthoritative           bool
	ReconciliationTriggeredAt *time.Time
	CreatedAt                 time.Time
}

type VCCRow struct {
	RowID             string
	VCCFileID         string
	VCCPaymentGroupID string
	LocationID        string
	PaymentID         string
	TraceID           string
	PayerName         string
	ProviderNPI       string
	ProviderTaxID     string
	IssueDate         string
	Amount            string
	CardFingerprint   string
	Last4             string
	ExpirationDate    string
	PatientID         string
	ClaimID           string
	ServiceDateStart  string
	ServiceDateEnd    string
}

// ---- Params types (pure Go; no pgtype.*) ----

type CreateERAParams struct {
	ID            string
	LocationID    string
	PayerName     string
	ProviderNPI   string
	ProviderTaxID string
	BPRAmount     string // decimal string, never float64
	PaymentMethod string
	TraceNumber   string
	FileHash      string
	StorageKey    string
	Status        string
}

type CreateVCCFileParams struct {
	ID             string
	LocationID     string
	FileHash       string
	StorageKey     string
	RowCount       int
	SourceFilename string
	Status         string
}

type UpdateERAParsedParams struct {
	ID            string
	PayerName     string
	ProviderNPI   string
	ProviderTaxID string
	BPRAmount     string
	PaymentMethod string
	TraceNumber   string
}

type UpdateVCCFileParsedParams struct {
	ID       string
	RowCount int
}

type InsertTransitionParams struct {
	ID         string
	EntityType string
	EntityID   string
	FromState  string
	ToState    string
	Reason     string
}

type ReconcileTriggerRequest struct {
	ERAGroupID string
	VCCGroupID string
}

// ---- Interfaces ----

// RemoteClient abstracts the seeder SDK for testability.
type RemoteClient interface {
	ListObjects(ctx context.Context, fileType string) (seedersdk.ListObjectsResponse, error)
	GetObject(ctx context.Context, key string) (seedersdk.GetObjectResponse, error)
}

// Decryptor decrypts file bytes using the wrapped data key and nonce from object metadata.
type Decryptor interface {
	Decrypt(encryptedBytes []byte, encryptedDataKey, nonce string) ([]byte, error)
}

// RawFileStore persists the original encrypted payload and returns its storage key.
type RawFileStore interface {
	Save(ctx context.Context, preferredKey, sourceKey string, encryptedBytes []byte) (string, error)
}

// Tokenizer converts raw card data into a processor-side payment method token.
// The implementation must not persist card data beyond the scope of the call.
type Tokenizer interface {
	CreatePaymentMethod(ctx context.Context, cardNumber, expMonth, expYear, cvv string) (string, error)
}

// IngesterStore is the persistence interface for all ingester operations.
// All methods use plain Go types (no pgtype.*) to keep mocks simple.
type IngesterStore interface {
	// Pre-decrypt dedup — uses PlaintextSHA256 as fileHash
	ExistsERAByOfficeAndHash(ctx context.Context, locationID, fileHash string) (bool, error)
	CreateERA(ctx context.Context, p CreateERAParams) error
	UpdateERAParsed(ctx context.Context, p UpdateERAParsedParams) error
	SetERAParseFailure(ctx context.Context, eraID string) error

	ExistsVCCByOfficeAndHash(ctx context.Context, locationID, fileHash string) (bool, error)
	CreateVCCFile(ctx context.Context, p CreateVCCFileParams) error
	UpdateVCCFileParsed(ctx context.Context, p UpdateVCCFileParsedParams) error
	SetVCCFileParseFailure(ctx context.Context, vccFileID string, rowCount int) error

	// ERA payment groups
	GetActiveERAPaymentGroup(ctx context.Context, locationID, traceNumber string) (*ERAPaymentGroup, error)
	CreateERAPaymentGroup(ctx context.Context, g *ERAPaymentGroup) error
	SetERAPaymentGroupException(ctx context.Context, groupID string) error
	MarkERAReconciliationTriggered(ctx context.Context, groupID string) error

	// VCC payment groups
	GetActiveVCCPaymentGroup(ctx context.Context, locationID, traceID, fingerprint string) (*VCCPaymentGroup, error)
	GetMatchingVCCForERA(ctx context.Context, locationID, traceNumber string) (*VCCPaymentGroup, error)
	GetMatchingERAForVCC(ctx context.Context, locationID, traceID string) (*ERAPaymentGroup, error)
	CreateVCCPaymentGroup(ctx context.Context, g *VCCPaymentGroup) error
	SetVCCPaymentGroupException(ctx context.Context, groupID string) error
	SupersedeVCCPaymentGroup(ctx context.Context, oldGroupID string, newGroup *VCCPaymentGroup) error
	MarkVCCReconciliationTriggered(ctx context.Context, groupID string) error

	// VCC rows
	InsertVCCRow(ctx context.Context, r *VCCRow) error

	// Audit trail
	InsertTransition(ctx context.Context, p InsertTransitionParams) error
}

// ReconcileTrigger emits a downstream reconciliation event.
type ReconcileTrigger interface {
	Trigger(ctx context.Context, req ReconcileTriggerRequest) error
}

// ---- Config ----

type Config struct {
	LocationID     string
	FingerprintKey string
	// ExpectedNPI is compared against the provider NPI parsed from every incoming ERA and
	// VCC file. Files whose NPI doesn't match are rejected as misrouted.
	//
	// This is populated at startup from the upstream API's GET /identity response. The
	// credential (token) scopes which practice's files the API returns; the identity
	// endpoint ties that credential to a specific NPI, giving the ingester a second layer
	// of defence against misrouted files even if the upstream delivers the wrong file.
	ExpectedNPI string
}

// ---- Service ----

type Service struct {
	cfg       Config
	client    RemoteClient
	decrypt   Decryptor
	rawStore  RawFileStore
	store     IngesterStore
	trigger   ReconcileTrigger
	tokenizer Tokenizer
	log       *slog.Logger
}

func NewService(
	cfg Config,
	client RemoteClient,
	dec Decryptor,
	rawStore RawFileStore,
	store IngesterStore,
	trigger ReconcileTrigger,
	tokenizer Tokenizer,
) *Service {
	return &Service{
		cfg:       cfg,
		client:    client,
		decrypt:   dec,
		rawStore:  rawStore,
		store:     store,
		trigger:   trigger,
		tokenizer: tokenizer,
		log:       slog.Default(),
	}
}

// FileError records a per-file processing failure. Run() accumulates these instead of
// halting early — one bad file must not block the rest (ING-001).
type FileError struct {
	Key    string
	Reason error
}

func (e *FileError) Error() string { return fmt.Sprintf("file %q: %v", e.Key, e.Reason) }

// RunResult summarises a completed Run.
type RunResult struct {
	Total      int
	Processed  int
	Duplicates int
	Errors     []*FileError
}

var errDuplicateFile = errors.New("duplicate file")

// Run polls the remote server, deduplicates, decrypts, parses, stores, and triggers reconciliation.
// All per-file errors are collected and returned in RunResult — a single bad file never blocks others.
func (s *Service) Run(ctx context.Context) (*RunResult, error) {
	resp, err := s.client.ListObjects(ctx, allFileTypesFilter)
	if err != nil {
		return nil, fmt.Errorf("ingester: list objects: %w", err)
	}

	result := &RunResult{Total: len(resp.Objects)}
	for _, obj := range resp.Objects {
		if err := s.processObject(ctx, obj); err != nil {
			if errors.Is(err, errDuplicateFile) {
				result.Duplicates++
				continue
			}
			result.Errors = append(result.Errors, &FileError{Key: obj.Key, Reason: err})
		} else {
			result.Processed++
		}
	}
	return result, nil
}

func (s *Service) processObject(ctx context.Context, obj seedersdk.ObjectMetadata) error {
	locationID := s.cfg.LocationID
	// The fileHash is a one way hash of the encrypted plaintext bytes.
	// It is a reliable value for detecting duplicates before decrypting
	fileHash := obj.Encryption.PlaintextSHA256

	// Pre-decrypt dedup — check by (locationID, plaintextHash) before downloading
	switch obj.FileType {
	case FileTypeERA:
		exists, err := s.store.ExistsERAByOfficeAndHash(ctx, locationID, fileHash)
		if err != nil {
			return err
		}
		if exists {
			return errDuplicateFile
		}
	case FileTypeVCC:
		exists, err := s.store.ExistsVCCByOfficeAndHash(ctx, locationID, fileHash)
		if err != nil {
			return err
		}
		if exists {
			return errDuplicateFile
		}
	default:
		return fmt.Errorf("unknown file type %q", obj.FileType)
	}

	// Download and decrypt
	res, err := s.client.GetObject(ctx, obj.Key)
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}

	storageKey, err := s.rawStore.Save(ctx, obj.RawStorageKey, obj.Key, res.EncryptedBytes)
	if err != nil {
		return fmt.Errorf("store raw object %q: %w", obj.Key, err)
	}

	switch obj.FileType {
	case FileTypeERA:
		eraID := newID()
		if err := s.store.CreateERA(ctx, CreateERAParams{
			ID:         eraID,
			LocationID: locationID,
			FileHash:   fileHash,
			StorageKey: storageKey,
			Status:     ERARemittanceStatusReceivedRaw,
		}); err != nil {
			return err
		}
		if err := s.store.InsertTransition(ctx, InsertTransitionParams{
			ID:         newID(),
			EntityType: EntityTypeERARemittance,
			EntityID:   eraID,
			ToState:    ERARemittanceStatusReceivedRaw,
		}); err != nil {
			return err
		}

		plaintext, err := s.decrypt.Decrypt(
			res.EncryptedBytes,
			res.Metadata.Encryption.EncryptedDataKey,
			res.Metadata.Encryption.Nonce,
		)
		if err != nil {
			return fmt.Errorf("decrypt object %q: %w", obj.Key, err)
		}
		return s.processERA(ctx, eraID, locationID, plaintext)
	case FileTypeVCC:
		vccFileID := newID()
		if err := s.store.CreateVCCFile(ctx, CreateVCCFileParams{
			ID:             vccFileID,
			LocationID:     locationID,
			FileHash:       fileHash,
			StorageKey:     storageKey,
			RowCount:       0,
			SourceFilename: obj.SourceFilename,
			Status:         VCCFileStatusReceivedRaw,
		}); err != nil {
			return err
		}
		if err := s.store.InsertTransition(ctx, InsertTransitionParams{
			ID: newID(), EntityType: EntityTypeVCCFile, EntityID: vccFileID,
			ToState: VCCFileStatusReceivedRaw,
		}); err != nil {
			return err
		}

		plaintext, err := s.decrypt.Decrypt(
			res.EncryptedBytes, 
			res.Metadata.Encryption.EncryptedDataKey, 
			res.Metadata.Encryption.Nonce,
		)
		if err != nil {
			return fmt.Errorf("decrypt object %q: %w", obj.Key, err)
		}
		return s.processVCC(ctx, vccFileID, locationID, plaintext)
	}

	return nil
}

func (s *Service) processERA(ctx context.Context, eraID, locationID string, plaintext []byte) error {
	parsed, parseErr := ParseERA(plaintext)

	if parseErr != nil {
		if err := s.store.SetERAParseFailure(ctx, eraID); err != nil {
			return err
		}
		if err := s.store.InsertTransition(ctx, InsertTransitionParams{
			ID: newID(), EntityType: EntityTypeERARemittance, EntityID: eraID,
			FromState: ERARemittanceStatusReceivedRaw,
			ToState:   ERARemittanceStatusExceptionParseFailed, Reason: parseErr.Error(),
		}); err != nil {
			return err
		}
		return nil
	}

	// ING-024: reject files whose provider NPI doesn't match the configured expected NPI.
	// This guards against misrouted files — the primary scope is the API credential, but
	// an explicit NPI check adds a second layer of defence.
	if s.cfg.ExpectedNPI != "" && parsed.ProviderNPI != s.cfg.ExpectedNPI {
		return fmt.Errorf("NPI mismatch: ERA contains provider NPI %q but this location expects %q — possible misrouted file",
			parsed.ProviderNPI, s.cfg.ExpectedNPI)
	}

	if err := s.store.UpdateERAParsed(ctx, UpdateERAParsedParams{
		ID:            eraID,
		PayerName:     parsed.PayerName,
		ProviderNPI:   parsed.ProviderNPI,
		ProviderTaxID: parsed.ProviderTaxID,
		BPRAmount:     parsed.BPRAmount,
		PaymentMethod: parsed.PaymentMethod,
		TraceNumber:   parsed.TraceNumber,
	}); err != nil {
		return err
	}
	if err := s.store.InsertTransition(ctx, InsertTransitionParams{
		ID:         newID(),
		EntityType: EntityTypeERARemittance,
		EntityID:   eraID,
		FromState:  ERARemittanceStatusReceivedRaw,
		ToState:    ERARemittanceStatusParsed,
	}); err != nil {
		return err
	}

	claimsJSON, adjustmentsJSON, err := marshalERAGroupJSON(parsed)
	if err != nil {
		return err
	}

	// ING-017: check for duplicate trace — partial unique index enforces one active group per trace
	existing, err := s.store.GetActiveERAPaymentGroup(ctx, locationID, parsed.TraceNumber)
	if err != nil {
		return err
	}
	if existing != nil {
		// Conflict: two distinct ERAs with same trace → both become EXCEPTION
		if err := s.store.SetERAPaymentGroupException(ctx, existing.GroupID); err != nil {
			return err
		}
		if err := s.store.InsertTransition(ctx, InsertTransitionParams{
			ID:         newID(),
			EntityType: EntityTypeERAPaymentGrp,
			EntityID:   existing.GroupID,
			FromState:  existing.Status,
			ToState:    ERAStatusException,
			Reason:     "duplicate trace number",
		}); err != nil {
			return err
		}
		exGroup := &ERAPaymentGroup{
			GroupID:     newID(),
			EraID:       eraID,
			LocationID:  locationID,
			TraceNumber: parsed.TraceNumber,
			BPRAmount:   parsed.BPRAmount,
			Status:      ERAStatusException,
		}
		if err := s.store.CreateERAPaymentGroup(ctx, exGroup); err != nil {
			return err
		}
		if err := s.store.InsertTransition(ctx, InsertTransitionParams{
			ID: newID(), EntityType: EntityTypeERAPaymentGrp, EntityID: exGroup.GroupID,
			ToState: ERAStatusException, Reason: "duplicate trace number",
		}); err != nil {
			return err
		}
		return nil
	}

	group := &ERAPaymentGroup{
		GroupID:     newID(),
		EraID:       eraID,
		LocationID:  locationID,
		TraceNumber: parsed.TraceNumber,
		BPRAmount:   parsed.BPRAmount,
		ClaimCount:  len(parsed.Claims),
		Claims:      claimsJSON,
		Adjustments: adjustmentsJSON,
		Status:      ERAStatusAwaitingVCC,
	}
	if err := s.store.CreateERAPaymentGroup(ctx, group); err != nil {
		return err
	}
	if err := s.store.InsertTransition(ctx, InsertTransitionParams{
		ID: newID(), EntityType: EntityTypeERAPaymentGrp, EntityID: group.GroupID,
		ToState: ERAStatusAwaitingVCC,
	}); err != nil {
		return err
	}

	// Check for a matching VCC group that arrived earlier
	matchingVCC, err := s.store.GetMatchingVCCForERA(ctx, locationID, parsed.TraceNumber)
	if err != nil {
		return err
	}
	if matchingVCC != nil {
		return s.triggerReconciliation(ctx, group, matchingVCC)
	}
	// No match yet — notify reconciler so it can check existing unmatched VCC groups (ING-012a)
	return s.trigger.Trigger(ctx, ReconcileTriggerRequest{ERAGroupID: group.GroupID})
}

func (s *Service) processVCC(ctx context.Context, vccFileID, locationID string, plaintext []byte) error {
	parsed, parseErr := ParseVCC(plaintext, s.cfg.FingerprintKey)

	rowCount := 0
	if parsed != nil {
		rowCount = len(parsed.Rows)
	}

	if parseErr != nil {
		if err := s.store.SetVCCFileParseFailure(ctx, vccFileID, rowCount); err != nil {
			return err
		}
		if err := s.store.InsertTransition(ctx, InsertTransitionParams{
			ID: newID(), EntityType: EntityTypeVCCFile, EntityID: vccFileID,
			FromState: VCCFileStatusReceivedRaw,
			ToState:   VCCFileStatusExceptionParseFailed, Reason: parseErr.Error(),
		}); err != nil {
			return err
		}
		return nil
	}

	// ING-024: reject if any row's NPI doesn't match the expected NPI for this location.
	if s.cfg.ExpectedNPI != "" {
		for _, row := range parsed.Rows {
			if row.ProviderNPI != s.cfg.ExpectedNPI {
				return fmt.Errorf("NPI mismatch: VCC row contains provider NPI %q but this location expects %q — possible misrouted file",
					row.ProviderNPI, s.cfg.ExpectedNPI)
			}
		}
	}

	if err := s.store.UpdateVCCFileParsed(ctx, UpdateVCCFileParsedParams{
		ID:       vccFileID,
		RowCount: rowCount,
	}); err != nil {
		return err
	}
	if err := s.store.InsertTransition(ctx, InsertTransitionParams{
		ID: newID(), EntityType: EntityTypeVCCFile, EntityID: vccFileID,
		FromState: VCCFileStatusReceivedRaw, ToState: VCCFileStatusParsed,
	}); err != nil {
		return err
	}

	for _, draftGroup := range parsed.Groups {
		if err := s.processVCCGroup(ctx, locationID, vccFileID, draftGroup); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) processVCCGroup(ctx context.Context, locationID, vccFileID string, draft *VCCPaymentGroupDraft) error {
	// VCC correction logic (PRD §9.11)
	existing, err := s.store.GetActiveVCCPaymentGroup(ctx, locationID, draft.TraceID, draft.CardFingerprint)
	if err != nil {
		return err
	}

	if existing != nil {
		if existing.TotalAmount == draft.TotalAmount {
			if existing.Status == VCCStatusAwaitingERA {
				// Case B: same fingerprint + same amount but group not yet processed → supersede
				pmID, err := s.tokenizeGroup(ctx, draft)
				if err != nil {
					return err
				}
				newGroup := s.buildVCCGroup(vccFileID, locationID, draft, pmID)
				if err := s.store.SupersedeVCCPaymentGroup(ctx, existing.GroupID, newGroup); err != nil {
					return err
				}
				if err := s.store.InsertTransition(ctx, InsertTransitionParams{
					ID: newID(), EntityType: EntityTypeVCCPaymentGrp, EntityID: newGroup.GroupID,
					ToState: VCCStatusAwaitingERA, Reason: "superseded prior authoritative version",
				}); err != nil {
					return err
				}
				return s.finalizeVCCGroup(ctx, locationID, newGroup, draft.Rows)
			}
			// Case A: already processed exact dup → silently ignore
			return nil
		}
		// Case C: different amount → EXCEPTION (skip tokenization; group will never be charged)
		exGroup := s.buildVCCGroup(vccFileID, locationID, draft, "")
		exGroup.Status = VCCStatusException
		if err := s.store.CreateVCCPaymentGroup(ctx, exGroup); err != nil {
			return err
		}
		if err := s.store.InsertTransition(ctx, InsertTransitionParams{
			ID: newID(), EntityType: EntityTypeVCCPaymentGrp, EntityID: exGroup.GroupID,
			ToState: VCCStatusException, Reason: "material conflict: amount mismatch",
		}); err != nil {
			return err
		}
		return nil
	}

	pmID, err := s.tokenizeGroup(ctx, draft)
	if err != nil {
		return err
	}
	newGroup := s.buildVCCGroup(vccFileID, locationID, draft, pmID)
	if err := s.store.CreateVCCPaymentGroup(ctx, newGroup); err != nil {
		return err
	}
	if err := s.store.InsertTransition(ctx, InsertTransitionParams{
		ID: newID(), EntityType: EntityTypeVCCPaymentGrp, EntityID: newGroup.GroupID,
		ToState: VCCStatusAwaitingERA,
	}); err != nil {
		return err
	}
	return s.finalizeVCCGroup(ctx, locationID, newGroup, draft.Rows)
}

// tokenizeGroup calls the tokenizer using the first row's card data, then clears card numbers
// from all rows in the draft. All rows in a group share the same card (same fingerprint).
func (s *Service) tokenizeGroup(ctx context.Context, draft *VCCPaymentGroupDraft) (string, error) {
	if len(draft.Rows) == 0 {
		return "", fmt.Errorf("tokenizeGroup: no rows in draft group")
	}
	if s.tokenizer == nil {
		return "", fmt.Errorf("tokenizeGroup: tokenizer is nil")
	}
	first := draft.Rows[0]
	cardNumber := first.cardNumber
	// Clear card numbers from all rows immediately after reading.
	for _, r := range draft.Rows {
		r.cardNumber = ""
	}
	expMonth, expYear, err := parseExpiration(first.ExpirationDate)
	if err != nil {
		return "", fmt.Errorf("parse expiration for group %s: %w", draft.TraceID, err)
	}
	pmID, err := s.tokenizer.CreatePaymentMethod(ctx, cardNumber, expMonth, expYear, "")
	if err != nil {
		return "", fmt.Errorf("tokenize card for group %s: %w", draft.TraceID, err)
	}
	return pmID, nil
}

// parseExpiration splits an expiration string (MM/YY, MM/YYYY, MMYY, MMYYYY) into month and year.
func parseExpiration(exp string) (month, year string, err error) {
	exp = strings.TrimSpace(exp)
	if exp == "" {
		return "", "", fmt.Errorf("expiration is empty")
	}

	if parts := strings.Split(exp, "/"); len(parts) == 2 {
		month, year = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	} else if parts := strings.Split(exp, "-"); len(parts) == 2 {
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		// Support both YYYY-MM (common in upstream CSV) and MM-YYYY / MM-YY.
		if len(left) == 4 {
			year, month = left, right
		} else {
			month, year = left, right
		}
	} else if len(exp) == 6 {
		month, year = exp[:2], exp[2:]
	} else if len(exp) == 4 {
		month, year = exp[:2], exp[2:]
	} else {
		return "", "", fmt.Errorf("unsupported expiration format %q", exp)
	}

	if len(year) == 2 {
		year = "20" + year
	}

	if len(month) == 1 {
		month = "0" + month
	}
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
		return "", "", fmt.Errorf("invalid expiration month %q", month)
	}

	if len(year) != 4 {
		return "", "", fmt.Errorf("invalid expiration year %q", year)
	}
	if _, err := strconv.Atoi(year); err != nil {
		return "", "", fmt.Errorf("invalid expiration year %q", year)
	}

	return month, year, nil
}

func (s *Service) finalizeVCCGroup(ctx context.Context, locationID string, group *VCCPaymentGroup, rows []*ParsedVCCRow) error {
	for _, row := range rows {
		if err := s.store.InsertVCCRow(ctx, &VCCRow{
			RowID: newID(), VCCFileID: group.VCCFileID, VCCPaymentGroupID: group.GroupID,
			LocationID: locationID, PaymentID: row.PaymentID, TraceID: row.TraceID,
			PayerName: row.PayerName, ProviderNPI: row.ProviderNPI, ProviderTaxID: row.ProviderTaxID,
			IssueDate: row.IssueDate, Amount: row.amountRaw,
			CardFingerprint: row.CardFingerprint, Last4: row.Last4,
			ExpirationDate: row.ExpirationDate, PatientID: row.PatientID, ClaimID: row.ClaimID,
			ServiceDateStart: row.ServiceDateStart, ServiceDateEnd: row.ServiceDateEnd,
		}); err != nil {
			return err
		}
	}

	matchingERA, err := s.store.GetMatchingERAForVCC(ctx, locationID, group.TraceID)
	if err != nil {
		return err
	}
	if matchingERA != nil {
		return s.triggerReconciliation(ctx, matchingERA, group)
	}
	return s.trigger.Trigger(ctx, ReconcileTriggerRequest{VCCGroupID: group.GroupID})
}

func marshalERAGroupJSON(parsed *ParsedERA) ([]byte, []byte, error) {
	claimsJSON, err := json.Marshal(parsed.Claims)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal ERA claims: %w", err)
	}
	adjustmentsJSON, err := json.Marshal(parsed.Adjustments)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal ERA adjustments: %w", err)
	}
	return claimsJSON, adjustmentsJSON, nil
}

func (s *Service) buildVCCGroup(vccFileID, locationID string, draft *VCCPaymentGroupDraft, paymentMethodID string) *VCCPaymentGroup {
	npi, taxID := "", ""
	if len(draft.Rows) > 0 {
		npi = draft.Rows[0].ProviderNPI
		taxID = draft.Rows[0].ProviderTaxID
	}
	return &VCCPaymentGroup{
		GroupID:         newID(),
		VCCFileID:       vccFileID,
		LocationID:      locationID,
		TraceID:         draft.TraceID,
		PaymentID:       draft.PaymentID,
		ProviderNPI:     npi,
		ProviderTaxID:   taxID,
		CardFingerprint: draft.CardFingerprint,
		TotalAmount:     draft.TotalAmount,
		PaymentMethodID: paymentMethodID,
		Status:          VCCStatusAwaitingERA,
		IsAuthoritative: true,
	}
}

// triggerReconciliation emits a downstream event exactly once (ING-023).
func (s *Service) triggerReconciliation(ctx context.Context, era *ERAPaymentGroup, vcc *VCCPaymentGroup) error {
	if era.ReconciliationTriggeredAt != nil || vcc.ReconciliationTriggeredAt != nil {
		return nil // already triggered — idempotent guard
	}
	if err := s.trigger.Trigger(ctx, ReconcileTriggerRequest{ERAGroupID: era.GroupID, VCCGroupID: vcc.GroupID}); err != nil {
		return fmt.Errorf("reconciliation trigger: %w", err)
	}
	if err := s.store.MarkERAReconciliationTriggered(ctx, era.GroupID); err != nil {
		return err
	}
	if err := s.store.MarkVCCReconciliationTriggered(ctx, vcc.GroupID); err != nil {
		return err
	}
	return nil
}

// newID generates a random hex ID.
func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

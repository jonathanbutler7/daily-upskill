package ingester

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ParsedVCCRow is an individual row from the VCC CSV, with PAN/CVV retained
// in memory only for fingerprinting and tokenization — never written to the DB.
type ParsedVCCRow struct {
	PaymentID     string
	TraceID       string
	PayerName     string // normalized
	ProviderNPI   string
	ProviderTaxID string
	IssueDate     string
	ExpirationDate string
	PatientID     string
	ClaimID       string
	ServiceDateStart string
	ServiceDateEnd   string

	// Derived — stored in DB instead of raw card data (ING-013)
	CardFingerprint string
	Last4           string

	// cardNumber is transient: used only for tokenization in processVCCGroup and
	// cleared immediately after. Never written to the DB.
	cardNumber string

	// Raw amount string (never float64) — used during grouping
	amountRaw string
}

// VCCPaymentGroupDraft is a group of rows sharing the same logical payment instrument.
// Groups are keyed by (trace_id, payment_id, provider_npi, provider_tax_id, card_fingerprint).
type VCCPaymentGroupDraft struct {
	TraceID         string
	PaymentID       string
	ProviderNPI     string
	ProviderTaxID   string
	CardFingerprint string
	TotalAmount     string // exact decimal, summed with math/big — never float64 (ING-021)
	Rows            []*ParsedVCCRow
}

// ParsedVCC is the result of parsing a VCC CSV file.
type ParsedVCC struct {
	Rows   []*ParsedVCCRow
	Groups []*VCCPaymentGroupDraft
}

var requiredVCCColumns = []string{
	"payment_id", "trace_id", "payer_name", "provider_npi", "provider_tax_id",
	"issue_date", "amount", "card_number", "expiration_date", "cvv",
	"patient_name", "patient_id", "claim_id", "service_date_start", "service_date_end",
}

// ParseVCC parses a VCC CSV plaintext into a ParsedVCC.
// fingerprintKey is used by CardFingerprint(). Raw PAN/CVV are never stored.
func ParseVCC(plaintext []byte, fingerprintKey string) (*ParsedVCC, error) {
	r := csv.NewReader(bytes.NewReader(plaintext))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("VCC parse error: invalid CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("VCC parse error: no data rows (only %d records)", len(records))
	}

	header := records[0]
	colIdx, err := buildColumnIndex(header, requiredVCCColumns)
	if err != nil {
		return nil, fmt.Errorf("VCC parse error: %w", err)
	}

	// groupKey → draft
	groups := map[string]*VCCPaymentGroupDraft{}
	var allRows []*ParsedVCCRow

	for i, record := range records[1:] {
		if len(record) != len(header) {
			return nil, fmt.Errorf("VCC parse error: row %d has %d columns, header has %d", i+1, len(record), len(header))
		}

		cardNumber := record[colIdx["card_number"]]
		amountStr  := strings.TrimSpace(record[colIdx["amount"]])

		if !isValidDecimal(amountStr) {
			return nil, fmt.Errorf("VCC parse error: row %d has invalid amount %q", i+1, amountStr)
		}

		fp   := CardFingerprint(cardNumber, fingerprintKey)
		last := Last4(cardNumber)

		issueDate := strings.TrimSpace(record[colIdx["issue_date"]])
		if _, err := time.Parse("2006-01-02", issueDate); err != nil {
			return nil, fmt.Errorf("VCC parse error: row %d has invalid issue_date %q", i+1, issueDate)
		}

		row := &ParsedVCCRow{
			PaymentID:        strings.TrimSpace(record[colIdx["payment_id"]]),
			TraceID:          strings.TrimSpace(record[colIdx["trace_id"]]),
			PayerName:        NormalizePayerName(record[colIdx["payer_name"]]), // ING-022
			ProviderNPI:      strings.TrimSpace(record[colIdx["provider_npi"]]),
			ProviderTaxID:    strings.TrimSpace(record[colIdx["provider_tax_id"]]),
			IssueDate:        issueDate,
			ExpirationDate:   strings.TrimSpace(record[colIdx["expiration_date"]]),
			PatientID:        strings.TrimSpace(record[colIdx["patient_id"]]),
			ClaimID:          strings.TrimSpace(record[colIdx["claim_id"]]),
			ServiceDateStart: strings.TrimSpace(record[colIdx["service_date_start"]]),
			ServiceDateEnd:   strings.TrimSpace(record[colIdx["service_date_end"]]),
			CardFingerprint:  fp,
			Last4:            last,
			cardNumber:       cardNumber, // transient; cleared after tokenization
			amountRaw:        amountStr,
			// CVV intentionally NOT stored in this struct
		}

		allRows = append(allRows, row)

		// Group by (trace_id, payment_id, provider_npi, provider_tax_id, card_fingerprint) — ING-016
		gKey := strings.Join([]string{row.TraceID, row.PaymentID, row.ProviderNPI, row.ProviderTaxID, fp}, "|")
		if _, exists := groups[gKey]; !exists {
			groups[gKey] = &VCCPaymentGroupDraft{
				TraceID:         row.TraceID,
				PaymentID:       row.PaymentID,
				ProviderNPI:     row.ProviderNPI,
				ProviderTaxID:   row.ProviderTaxID,
				CardFingerprint: fp,
			}
		}
		groups[gKey].Rows = append(groups[gKey].Rows, row)
	}

	// Sum amounts using math/big.Rat — never float64 (ING-021)
	var draftGroups []*VCCPaymentGroupDraft
	for _, g := range groups {
		total := new(big.Rat)
		for _, row := range g.Rows {
			amt := new(big.Rat)
			if _, ok := amt.SetString(row.amountRaw); !ok {
				return nil, fmt.Errorf("VCC parse error: could not parse amount %q as rational", row.amountRaw)
			}
			total.Add(total, amt)
		}
		g.TotalAmount = total.FloatString(2)
		draftGroups = append(draftGroups, g)
	}

	if len(draftGroups) == 0 {
		return nil, fmt.Errorf("VCC parse error: no payment groups derived from file")
	}

	return &ParsedVCC{Rows: allRows, Groups: draftGroups}, nil
}

func buildColumnIndex(header []string, required []string) (map[string]int, error) {
	idx := map[string]int{}
	for i, col := range header {
		idx[strings.TrimSpace(strings.ToLower(col))] = i
	}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return nil, fmt.Errorf("missing required column %q", col)
		}
	}
	return idx, nil
}

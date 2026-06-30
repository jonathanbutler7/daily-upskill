package ingester

import (
	"strings"
	"testing"
)

const baseVCCContent = `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end
PMT-A,9876543210,DELTA DENTAL OF CALIFORNIA,1234567890,12-3456789,2026-05-24,250.00,4111111111111111,2028-01,123,John Doe,PAT-001,CLM-001,2026-04-01,2026-04-01
PMT-A,9876543210,DELTA DENTAL OF CALIFORNIA,1234567890,12-3456789,2026-05-24,200.00,4111111111111111,2028-01,123,Jane Smith,PAT-002,CLM-002,2026-04-02,2026-04-02`

const multiTraceVCCContent = `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end
PMT-A,9876543210,DELTA DENTAL OF CALIFORNIA,1234567890,12-3456789,2026-05-24,450.00,4111111111111111,2028-01,123,John Doe,PAT-001,CLM-001,2026-04-01,2026-04-01
PMT-B,1111111111,AETNA,1234567890,12-3456789,2026-05-24,300.00,4222222222222222,2029-06,456,Jane Smith,PAT-002,CLM-002,2026-04-15,2026-04-15`

const decimalPrecisionVCCContent = `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end
PMT-A,9876543210,AETNA,1234567890,12-3456789,2026-05-24,100.10,4111111111111111,2028-01,123,John Doe,PAT-001,CLM-001,2026-04-01,2026-04-01
PMT-A,9876543210,AETNA,1234567890,12-3456789,2026-05-24,100.10,4111111111111111,2028-01,123,Jane Smith,PAT-002,CLM-002,2026-04-02,2026-04-02`

const testFingerprintKey = "test-hmac-key"

func TestParseVCC_HappyPath(t *testing.T) {
	parsed, err := ParseVCC([]byte(baseVCCContent), testFingerprintKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Rows) != 2 {
		t.Errorf("Rows count = %d, want 2", len(parsed.Rows))
	}
	if len(parsed.Groups) != 1 {
		t.Errorf("Groups count = %d, want 1 (same card, same trace)", len(parsed.Groups))
	}

	g := parsed.Groups[0]
	if g.TraceID != "9876543210" {
		t.Errorf("Group TraceID = %q, want 9876543210", g.TraceID)
	}
	if g.TotalAmount != "450.00" {
		t.Errorf("Group TotalAmount = %q, want 450.00", g.TotalAmount)
	}
}

func TestParseVCC_MultiTrace_ProducesIndependentGroups(t *testing.T) {
	// ING-016: two trace IDs in one file → two independent groups
	parsed, err := ParseVCC([]byte(multiTraceVCCContent), testFingerprintKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Groups) != 2 {
		t.Errorf("Groups count = %d, want 2 for two distinct trace IDs", len(parsed.Groups))
	}

	traces := map[string]string{}
	for _, g := range parsed.Groups {
		traces[g.TraceID] = g.TotalAmount
	}
	if traces["9876543210"] != "450.00" {
		t.Errorf("Group for trace 9876543210 amount = %q, want 450.00", traces["9876543210"])
	}
	if traces["1111111111"] != "300.00" {
		t.Errorf("Group for trace 1111111111 amount = %q, want 300.00", traces["1111111111"])
	}
}

func TestParseVCC_DecimalPrecision(t *testing.T) {
	// ING-021: 100.10 + 100.10 must equal exactly "200.20", not "200.19999..."
	parsed, err := ParseVCC([]byte(decimalPrecisionVCCContent), testFingerprintKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Groups) != 1 {
		t.Fatalf("Groups count = %d, want 1", len(parsed.Groups))
	}
	if parsed.Groups[0].TotalAmount != "200.20" {
		t.Errorf("TotalAmount = %q, want exactly 200.20 (decimal precision ING-021)", parsed.Groups[0].TotalAmount)
	}
	// Confirm it is NOT the float64 result
	if strings.HasPrefix(parsed.Groups[0].TotalAmount, "200.19") || strings.HasPrefix(parsed.Groups[0].TotalAmount, "200.20000") {
		t.Errorf("TotalAmount has float64 imprecision: %q", parsed.Groups[0].TotalAmount)
	}
}

func TestParseVCC_PANAndCVVNotInRows(t *testing.T) {
	// ING-013: raw PAN and CVV must not appear in any stored field
	parsed, err := ParseVCC([]byte(baseVCCContent), testFingerprintKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, row := range parsed.Rows {
		if row.CardFingerprint == "4111111111111111" {
			t.Errorf("row %d: CardFingerprint contains raw PAN", i)
		}
		// Verify fingerprint looks like our HMAC output (prefixed with "fp:")
		if !strings.HasPrefix(row.CardFingerprint, "fp:") {
			t.Errorf("row %d: CardFingerprint %q is not in expected fp: format", i, row.CardFingerprint)
		}
		if row.Last4 != "1111" {
			t.Errorf("row %d: Last4 = %q, want 1111", i, row.Last4)
		}
	}
}

func TestParseVCC_PayerNameNormalizedAtParseTime(t *testing.T) {
	// ING-022: payer name should be normalized before any storage
	content := `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end
PMT-A,1111,"  delta  dental  of california.  ",1234567890,12-3456789,2026-05-24,100.00,4111111111111111,2028-01,123,John Doe,PAT-001,CLM-001,2026-04-01,2026-04-01`
	parsed, err := ParseVCC([]byte(content), testFingerprintKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Rows[0].PayerName != "DELTA DENTAL OF CALIFORNIA" {
		t.Errorf("PayerName = %q, want normalized DELTA DENTAL OF CALIFORNIA", parsed.Rows[0].PayerName)
	}
}

func TestParseVCC_MissingRequiredColumn(t *testing.T) {
	content := `payment_id,trace_id,payer_name
PMT-A,1111,AETNA`
	_, err := ParseVCC([]byte(content), testFingerprintKey)
	if err == nil {
		t.Fatal("expected error for missing required columns, got nil")
	}
}

func TestParseVCC_InvalidAmount(t *testing.T) {
	content := `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end
PMT-A,1111,AETNA,1234567890,12-3456789,2026-05-24,NOTANUMBER,4111111111111111,2028-01,123,John Doe,PAT-001,CLM-001,2026-04-01,2026-04-01`
	_, err := ParseVCC([]byte(content), testFingerprintKey)
	if err == nil {
		t.Fatal("expected error for invalid amount, got nil")
	}
}

func TestParseVCC_NoDataRows(t *testing.T) {
	content := `payment_id,trace_id,payer_name,provider_npi,provider_tax_id,issue_date,amount,card_number,expiration_date,cvv,patient_name,patient_id,claim_id,service_date_start,service_date_end`
	_, err := ParseVCC([]byte(content), testFingerprintKey)
	if err == nil {
		t.Fatal("expected error for header-only CSV, got nil")
	}
}

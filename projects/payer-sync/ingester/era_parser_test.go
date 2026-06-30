package ingester

import "testing"

const testERAContent = `ISA*00*          *00*          *ZZ*DELTADENTALCA   *ZZ*ACMEDENTAL      *260524*1200*^*00501*000000001*0*P*:~
GS*HP*DELTADENTALCA*ACMEDENTAL*20260524*1200*1*X*005010X221A1~
ST*835*0001~
BPR*I*450.00*C*VCC~
TRN*1*9876543210*9876543210~
N1*PR*  Delta  Dental of California.  ~
N1*PE*ACME DENTAL GROUP*XX*1234567890~
REF*TJ*12-3456789~
CLP*CLM-001*1*250.00*250.00~
CAS*CO*45*15.00*1~
SVC*HC:D0120*250.00*250.00~
CLP*CLM-002*1*200.00*200.00~
SVC*HC:D0140*200.00*200.00~
SE*11*0001~
GE*1*1~
IEA*1*000000001~`

func TestParseERA_HappyPath(t *testing.T) {
	parsed, err := ParseERA([]byte(testERAContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.TraceNumber != "9876543210" {
		t.Errorf("TraceNumber = %q, want %q", parsed.TraceNumber, "9876543210")
	}
	// ING-022: payer name normalized at parse time
	if parsed.PayerName != "DELTA DENTAL OF CALIFORNIA" {
		t.Errorf("PayerName = %q, want %q", parsed.PayerName, "DELTA DENTAL OF CALIFORNIA")
	}
	if parsed.ProviderNPI != "1234567890" {
		t.Errorf("ProviderNPI = %q, want %q", parsed.ProviderNPI, "1234567890")
	}
	if parsed.ProviderTaxID != "12-3456789" {
		t.Errorf("ProviderTaxID = %q, want %q", parsed.ProviderTaxID, "12-3456789")
	}
	if parsed.BPRAmount != "450.00" {
		t.Errorf("BPRAmount = %q, want %q", parsed.BPRAmount, "450.00")
	}
	if parsed.PaymentMethod != "VCC" {
		t.Errorf("PaymentMethod = %q, want %q", parsed.PaymentMethod, "VCC")
	}
	if len(parsed.Claims) != 2 {
		t.Errorf("Claims count = %d, want 2", len(parsed.Claims))
	}
	if parsed.Claims[0].ClaimID != "CLM-001" {
		t.Errorf("Claims[0].ClaimID = %q, want CLM-001", parsed.Claims[0].ClaimID)
	}
	if len(parsed.Claims[0].ServiceLines) != 1 || parsed.Claims[0].ServiceLines[0].ProcedureCode != "HC:D0120" {
		t.Errorf("Claims[0].ServiceLines = %+v, want one HC:D0120 line", parsed.Claims[0].ServiceLines)
	}
	if len(parsed.Claims[0].Adjustments) != 1 || parsed.Claims[0].Adjustments[0].ReasonCode != "45" {
		t.Errorf("Claims[0].Adjustments = %+v, want one CAS adjustment with reason 45", parsed.Claims[0].Adjustments)
	}
	if parsed.Claims[1].ClaimID != "CLM-002" {
		t.Errorf("Claims[1].ClaimID = %q, want CLM-002", parsed.Claims[1].ClaimID)
	}
	if len(parsed.Adjustments) != 1 || parsed.Adjustments[0].Amount != "15.00" {
		t.Errorf("Adjustments = %+v, want one top-level adjustment amount 15.00", parsed.Adjustments)
	}
}

func TestParseERA_MissingTRN(t *testing.T) {
	content := `ISA*00*~
BPR*I*450.00*C*VCC~
N1*PR*DELTA DENTAL~
SE*3*0001~`
	_, err := ParseERA([]byte(content))
	if err == nil {
		t.Fatal("expected error for missing TRN, got nil")
	}
}

func TestParseERA_NonNumericBPRAmount(t *testing.T) {
	// ING-007: non-numeric BPR amount should trigger parse failure
	content := `ISA*00*~
BPR*I*NOTANUMBER*C*VCC~
TRN*1*9876543210~
SE*3*0001~`
	_, err := ParseERA([]byte(content))
	if err == nil {
		t.Fatal("expected error for non-numeric BPR amount, got nil")
	}
}

func TestParseERA_PayerNameNormalizedAtParseTime(t *testing.T) {
	// ING-022: normalization happens at parse time, not deferred
	content := `BPR*I*100.00*C*VCC~
TRN*1*1111~
N1*PR*  delta  dental, of california.  ~
SE*3*0001~`
	parsed, err := ParseERA([]byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.PayerName != "DELTA DENTAL OF CALIFORNIA" {
		t.Errorf("PayerName = %q, want normalized form", parsed.PayerName)
	}
}

func TestParseERA_MissingBPR(t *testing.T) {
	content := `TRN*1*9876543210~
N1*PR*DELTA DENTAL~
SE*2*0001~`
	_, err := ParseERA([]byte(content))
	if err == nil {
		t.Fatal("expected error for missing BPR, got nil")
	}
}

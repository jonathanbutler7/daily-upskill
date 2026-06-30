package ingester

import (
	"fmt"
	"strings"
)

// ParsedERA is the result of parsing an EDI X12 835 file.
type ParsedERA struct {
	TraceNumber   string
	PayerName     string // normalized via NormalizePayerName
	ProviderNPI   string
	ProviderTaxID string
	BPRAmount     string // exact decimal string — never float64 (ING-021)
	PaymentMethod string
	Claims        []ParsedERAClaim
	Adjustments   []ParsedERAAdjustment
}

// ParsedERAClaim is a single CLP segment extracted from the ERA.
type ParsedERAClaim struct {
	ClaimID      string
	Amount       string
	ServiceLines []ParsedERAServiceLine
	Adjustments  []ParsedERAAdjustment
}

type ParsedERAServiceLine struct {
	ProcedureCode string
	BilledAmount  string
	PaidAmount    string
}

type ParsedERAAdjustment struct {
	GroupCode  string
	ReasonCode string
	Amount     string
	Quantity   string
}

// ParseERA parses an EDI X12 835 plaintext byte slice into a ParsedERA.
// Segments are separated by '~', elements by '*'.
// Returns an error if required fields (TRN, BPR amount) are missing or malformed.
func ParseERA(plaintext []byte) (*ParsedERA, error) {
	content := string(plaintext)
	// Normalise line endings — some generators include \n after ~
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, "\r", "")

	segments := strings.Split(content, "~")

	result := &ParsedERA{}
	inPEContext := false // tracks whether we are inside an N1*PE provider loop
	var currentClaim *ParsedERAClaim

	for _, raw := range segments {
		seg := strings.TrimSpace(raw)
		if seg == "" {
			continue
		}
		elems := strings.Split(seg, "*")
		if len(elems) == 0 {
			continue
		}

		switch elems[0] {
		case "BPR":
			// BPR*I*<amount>*C*<payment_method>*...
			if len(elems) < 5 {
				return nil, fmt.Errorf("ERA parse error: BPR segment too short: %q", seg)
			}
			amount := strings.TrimSpace(elems[2])
			if !isValidDecimal(amount) {
				return nil, fmt.Errorf("ERA parse error: BPR amount is not a valid decimal: %q", amount)
			}
			result.BPRAmount = amount
			result.PaymentMethod = strings.TrimSpace(elems[4])

		case "TRN":
			// TRN*1*<trace_number>*...
			if len(elems) < 3 {
				return nil, fmt.Errorf("ERA parse error: TRN segment too short: %q", seg)
			}
			result.TraceNumber = strings.TrimSpace(elems[2])

		case "N1":
			if len(elems) < 3 {
				continue
			}
			qualifier := strings.TrimSpace(elems[1])
			switch qualifier {
			case "PR": // payer
				result.PayerName = NormalizePayerName(elems[2])
				inPEContext = false
			case "PE": // provider/payee
				inPEContext = true
				// N1*PE*<name>*XX*<NPI>
				if len(elems) >= 5 && strings.TrimSpace(elems[3]) == "XX" {
					result.ProviderNPI = strings.TrimSpace(elems[4])
				}
			default:
				inPEContext = false
			}

		case "REF":
			// REF*TJ*<tax_id> — provider tax ID, only valid in PE context
			if inPEContext && len(elems) >= 3 && strings.TrimSpace(elems[1]) == "TJ" {
				result.ProviderTaxID = strings.TrimSpace(elems[2])
			}

		case "CLP":
			// CLP*<claim_id>*<status>*<billed>*<paid>*...
			if len(elems) < 5 {
				continue
			}
			claimID := strings.TrimSpace(elems[1])
			amount := strings.TrimSpace(elems[4])
			if claimID != "" {
				result.Claims = append(result.Claims, ParsedERAClaim{
					ClaimID: claimID,
					Amount:  amount,
				})
				currentClaim = &result.Claims[len(result.Claims)-1]
			}

		case "SVC":
			if currentClaim == nil || len(elems) < 4 {
				continue
			}
			currentClaim.ServiceLines = append(currentClaim.ServiceLines, ParsedERAServiceLine{
				ProcedureCode: strings.TrimSpace(elems[1]),
				BilledAmount:  strings.TrimSpace(elems[2]),
				PaidAmount:    strings.TrimSpace(elems[3]),
			})

		case "CAS":
			if currentClaim == nil || len(elems) < 4 {
				continue
			}
			groupCode := strings.TrimSpace(elems[1])
			for i := 2; i+1 < len(elems); i += 3 {
				adj := ParsedERAAdjustment{
					GroupCode:  groupCode,
					ReasonCode: strings.TrimSpace(elems[i]),
					Amount:     strings.TrimSpace(elems[i+1]),
				}
				if i+2 < len(elems) {
					adj.Quantity = strings.TrimSpace(elems[i+2])
				}
				currentClaim.Adjustments = append(currentClaim.Adjustments, adj)
				result.Adjustments = append(result.Adjustments, adj)
			}
		}
	}

	if result.TraceNumber == "" {
		return nil, fmt.Errorf("ERA parse error: missing TRN (trace number)")
	}
	if result.BPRAmount == "" {
		return nil, fmt.Errorf("ERA parse error: missing BPR amount")
	}

	return result, nil
}

// isValidDecimal checks whether s is a non-empty numeric string with optional decimal point.
func isValidDecimal(s string) bool {
	if s == "" {
		return false
	}
	dotSeen := false
	for i, c := range s {
		switch {
		case c >= '0' && c <= '9':
			// ok
		case c == '.' && !dotSeen:
			dotSeen = true
		case c == '-' && i == 0:
			// negative amounts technically allowed
		default:
			return false
		}
	}
	return true
}

package server

import (
	"context"
	"math"
	"strings"
	"testing"

	willtheypay "weavelab.xyz/schema-gen-go/schemas/payments-platform/will-they-pay/v1"
)

func TestGetPaymentLikelihoodRequiresIDs(t *testing.T) {
	srv := NewWillTheyPayServer("v0-test")

	_, err := srv.GetPaymentLikelihood(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected error for nil request")
	}
}

func TestGetPaymentLikelihoodFromCSV(t *testing.T) {
	srv := NewWillTheyPayServer("v0-test")

	resp, err := srv.GetPaymentLikelihood(context.Background(), &willtheypay.GetPaymentLikelihoodRequest{
		PatientId:  "11af2c63-9a57-556a-91e7-c157840adf45",
		LocationId: "loc-1",
	})
	if err != nil {
		t.Fatalf("GetPaymentLikelihood() unexpected error: %v", err)
	}

	if resp.GetLikelihood() == nil {
		t.Fatalf("expected likelihood in response")
	}

	if resp.GetLikelihood().GetWillPayIn_30() {
		t.Fatalf("expected will_pay_in_30=false from CSV")
	}

	if resp.GetLikelihood().GetRiskBand() != "medium" {
		t.Fatalf("expected risk_band=medium from latest CSV row, got %q", resp.GetLikelihood().GetRiskBand())
	}

	if resp.GetLikelihood().GetModelVersion() != "will_they_pay_prod_v1" {
		t.Fatalf("expected model version from CSV, got %q", resp.GetLikelihood().GetModelVersion())
	}

	if math.Abs(float64(resp.GetLikelihood().GetConfidenceScore()-float32(0.4549343))) > 1e-8 {
		t.Fatalf("expected confidence_score from latest CSV row, got %f", resp.GetLikelihood().GetConfidenceScore())
	}
}

func TestGetPaymentLikelihoodMissingPrediction(t *testing.T) {
	srv := NewWillTheyPayServer("v0-test")

	_, err := srv.GetPaymentLikelihood(context.Background(), &willtheypay.GetPaymentLikelihoodRequest{
		PatientId:  "person-does-not-exist",
		LocationId: "loc-1",
	})
	if err == nil {
		t.Fatalf("expected error when prediction is missing")
	}

	if !strings.Contains(err.Error(), "no prediction found") {
		t.Fatalf("expected missing prediction error, got %v", err)
	}
}

func TestDeriveRiskBand(t *testing.T) {
	cases := []struct {
		score float32
		want  string
	}{
		{score: 0.70, want: "high"},
		{score: 0.55, want: "medium"},
		{score: 0.10, want: "low"},
	}

	for _, tc := range cases {
		got := deriveRiskBand(tc.score)
		if got != tc.want {
			t.Fatalf("deriveRiskBand(%f) = %q, want %q", tc.score, got, tc.want)
		}
	}
}

package ledger

import (
	"ledger-db/internal/ledgerstore"
	"testing"
)

func TestReversalRequestValidation(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name           string
		transactionID  ledgerstore.TransactionID
		reason         ledgerstore.Reason
		idempotencyKey ledgerstore.IdempotencyKey
		wantErr        error
	}{
		{
			name:           "transaction id is required",
			transactionID:  0,
			reason:         "duplicate transfer",
			idempotencyKey: "reverse-transfer",
			wantErr:        ledgerstore.ErrTransactionIDRequired,
		},
		{
			name:           "reason is required",
			transactionID:  1,
			reason:         "",
			idempotencyKey: "reverse-transfer",
			wantErr:        ledgerstore.ErrReasonIsRequired,
		},
		{
			name:           "reason cannot be blank",
			transactionID:  1,
			reason:         "   ",
			idempotencyKey: "reverse-transfer",
			wantErr:        ledgerstore.ErrReasonIsRequired,
		},
		{
			name:           "idempotency key is required",
			transactionID:  1,
			reason:         "duplicate transfer",
			idempotencyKey: "",
			wantErr:        ledgerstore.ErrIdempotencyKeyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Reversal(
				ctx,
				nil,
				ledgerstore.ReversalCommand{
					TransactionID:  tt.transactionID,
					Reason:         tt.reason,
					IdempotencyKey: tt.idempotencyKey,
				},
			)
			if err == nil {
				t.Fatal("expected error")
			}
			if err != tt.wantErr {
				t.Fatalf("got %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

package ledger

import (
	"testing"
)

func TestAddBalance(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name                string
		toAccountID         int64
		amount              int64
		rail                string
		externalReferenceID string
		idempotencyKey      string
		wantErr             string
	}{
		{
			name:                "",
			toAccountID:         0,
			amount:              0,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			wantErr:             "to account id is required",
		},
	}

	for _, tt := range tests {
		_, err := AddBalance(
			ctx, nil, tt.toAccountID, tt.amount, tt.rail, tt.externalReferenceID, tt.idempotencyKey,
		)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != tt.wantErr {
			t.Fatalf("got %q, want %q", err.Error(), tt.wantErr)
		}
	}
}

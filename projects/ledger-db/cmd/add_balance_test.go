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
		wantErr             error
	}{
		{
			name:                "to account id is required",
			toAccountID:         0,
			amount:              100,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			wantErr:             ErrToAccountIDRequired,
		},
		{
			name:                "transfer amount is required",
			toAccountID:         1,
			amount:              0,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			wantErr:             ErrTransferAmountRequired,
		},
		{
			name:                "rail value is required",
			toAccountID:         1,
			amount:              100,
			rail:                "",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			wantErr:             ErrRailValueRequired,
		},
		{
			name:                "external reference id is required",
			toAccountID:         1,
			amount:              100,
			rail:                "string",
			externalReferenceID: "",
			idempotencyKey:      "string",
			wantErr:             ErrExternalReferenceIdRequired,
		},
		{
			name:                "idempotency key is required",
			toAccountID:         1,
			amount:              100,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "",
			wantErr:             ErrIdempotencyKeyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AddBalance(
				ctx, nil, tt.toAccountID, tt.amount, tt.rail, tt.externalReferenceID, tt.idempotencyKey,
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

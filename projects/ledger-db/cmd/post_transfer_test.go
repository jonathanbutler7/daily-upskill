package ledger

import (
	"testing"
)

func TestPostTransfer(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name           string
		fromAccountID  int64
		toAccountID    int64
		amount         int64
		idempotencyKey string
		wantErr        error
	}{
		{
			name:           "from account id is required",
			fromAccountID:  0,
			toAccountID:    2,
			amount:         100,
			idempotencyKey: "string",
			wantErr:        ErrFromAccountIDRequired,
		},
		{
			name:           "to account id is required",
			fromAccountID:  1,
			toAccountID:    0,
			amount:         100,
			idempotencyKey: "string",
			wantErr:        ErrToAccountIDRequired,
		},
		{
			name:           "amount must be greater than zero",
			fromAccountID:  1,
			toAccountID:    2,
			amount:         0,
			idempotencyKey: "string",
			wantErr:        ErrAmountGreaterThanZero,
		},
		{
			name:           "idempotency key is required",
			fromAccountID:  1,
			toAccountID:    2,
			amount:         100,
			idempotencyKey: "",
			wantErr:        ErrIdempotencyKeyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PostTransfer(
				ctx, nil, tt.fromAccountID, tt.toAccountID, tt.amount, tt.idempotencyKey,
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

package ledger

import (
	"testing"
)

func TestPostTransfer(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		fromAccountID  int64
		toAccountID    int64
		amount         int64
		idempotencyKey string
		wantErr        error
	}{
		{
			fromAccountID:  0,
			toAccountID:    0,
			amount:         0,
			idempotencyKey: "string",
			wantErr:        ErrFromAccountIDRequired,
		},
	}

	for _, tt := range tests {
		_, err := PostTransfer(
			ctx, nil, tt.fromAccountID, tt.toAccountID, tt.amount, tt.idempotencyKey,
		)
		if err == nil {
			t.Fatal("expected error")
		}
		if err != tt.wantErr {
			t.Fatalf("got %q, want %q", err.Error(), tt.wantErr)
		}
	}
}

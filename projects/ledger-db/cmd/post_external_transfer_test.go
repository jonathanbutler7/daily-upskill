package ledger

import (
	"ledger-db/internal/ledgerstore"
	"testing"
)

func TestDepositFundsRequestValidation(t *testing.T) {
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
			wantErr:             ledgerstore.ErrToAccountIDRequired,
		},
		{
			name:                "transfer amount is required",
			toAccountID:         1,
			amount:              0,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			wantErr:             ledgerstore.ErrTransferAmountRequired,
		},
		{
			name:                "rail value is required",
			toAccountID:         1,
			amount:              100,
			rail:                "",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			wantErr:             ledgerstore.ErrRailValueRequired,
		},
		{
			name:                "external reference id is required",
			toAccountID:         1,
			amount:              100,
			rail:                "string",
			externalReferenceID: "",
			idempotencyKey:      "string",
			wantErr:             ledgerstore.ErrExternalReferenceRequired,
		},
		{
			name:                "idempotency key is required",
			toAccountID:         1,
			amount:              100,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "",
			wantErr:             ledgerstore.ErrIdempotencyKeyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PostExternalTransfer(
				ctx,
				nil,
				ledgerstore.PostExternalTransferCommand{
					ToAccountID:       ledgerstore.AccountID(tt.toAccountID),
					TransferAmount:    ledgerstore.Amount(tt.amount),
					Rail:              ledgerstore.PaymentRail(tt.rail),
					ExternalReference: ledgerstore.ExternalReference(tt.externalReferenceID),
					IdempotencyKey:    ledgerstore.IdempotencyKey(tt.idempotencyKey),
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

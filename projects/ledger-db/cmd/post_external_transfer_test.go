package ledger

import (
	"ledger-db/internal/ledgerstore"
	"ledger-db/internal/testdb"
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
		direction           ledgerstore.ExternalTransferDirection
		wantErr             error
	}{
		{
			name:                "to account id is required",
			toAccountID:         0,
			amount:              100,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			direction:           ledgerstore.ExternalTransferDirectionDeposit,
			wantErr:             ledgerstore.ErrToAccountIDRequired,
		},
		{
			name:                "transfer amount is required",
			toAccountID:         1,
			amount:              0,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			direction:           ledgerstore.ExternalTransferDirectionDeposit,
			wantErr:             ledgerstore.ErrTransferAmountRequired,
		},
		{
			name:                "rail value is required",
			toAccountID:         1,
			amount:              100,
			rail:                "",
			externalReferenceID: "string",
			idempotencyKey:      "string",
			direction:           ledgerstore.ExternalTransferDirectionDeposit,
			wantErr:             ledgerstore.ErrRailValueRequired,
		},
		{
			name:                "external reference id is required",
			toAccountID:         1,
			amount:              100,
			rail:                "string",
			externalReferenceID: "",
			idempotencyKey:      "string",
			direction:           ledgerstore.ExternalTransferDirectionDeposit,
			wantErr:             ledgerstore.ErrExternalReferenceRequired,
		},
		{
			name:                "idempotency key is required",
			toAccountID:         1,
			amount:              100,
			rail:                "string",
			externalReferenceID: "string",
			idempotencyKey:      "",
			direction:           ledgerstore.ExternalTransferDirectionDeposit,
			wantErr:             ledgerstore.ErrIdempotencyKeyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PostExternalTransfer(
				ctx,
				nil,
				ledgerstore.PostExternalTransferCommand{
					UserAccountID:             ledgerstore.AccountID(tt.toAccountID),
					TransferAmount:            ledgerstore.Amount(tt.amount),
					Rail:                      ledgerstore.PaymentRail(tt.rail),
					ExternalReference:         ledgerstore.ExternalReference(tt.externalReferenceID),
					IdempotencyKey:            ledgerstore.IdempotencyKey(tt.idempotencyKey),
					ExternalTransferDirection: ledgerstore.ExternalTransferDirectionDeposit,
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

func TestPostExternalTransferForwardsDeferredSettlementMode(t *testing.T) {
	ctx, db := testdb.OpenIntegrationDB(t)
	testdb.ResetIntegrationDB(t, ctx, db)

	var userAccountID int64
	err := db.QueryRowContext(ctx, `
		insert into ledger_accounts (name, description, currency_code, normal_balance, ledgerable_type, balance)
		values ('Deferred test wallet', 'Deferred test wallet', 'USD', 'credit', 'internal_account', 0)
		returning id;
	`).Scan(&userAccountID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = PostExternalTransfer(ctx, db, ledgerstore.PostExternalTransferCommand{
		UserAccountID:             ledgerstore.AccountID(userAccountID),
		TransferAmount:            100,
		Rail:                      "ach",
		ExternalReference:         "deferred-mode-forwarding-test",
		IdempotencyKey:            "deferred-mode-forwarding-test",
		ExternalTransferDirection: ledgerstore.ExternalTransferDirectionDeposit,
		SettlementBalanceMode:     ledgerstore.SettlementBalanceModeDeferredAsync,
	})
	if err != nil {
		t.Fatal(err)
	}

	var userBalance, settlementBalance int64
	if err := db.QueryRowContext(ctx, `select balance from ledger_accounts where id = $1`, userAccountID).Scan(&userBalance); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `select balance from ledger_accounts where name = 'Cash Settlement' and currency_code = 'USD'`).Scan(&settlementBalance); err != nil {
		t.Fatal(err)
	}
	if userBalance != 100 || settlementBalance != 0 {
		t.Fatalf("balances: user=%d settlement=%d, want 100 and 0", userBalance, settlementBalance)
	}
}

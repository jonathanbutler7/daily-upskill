package ledgerstore

import (
	"os"
	"strings"
	"testing"
)

func TestPostExternalTransferEntryAndBalanceDirectionsMatch(t *testing.T) {
	sourceBytes, err := os.ReadFile("external_transfer.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)

	tests := []struct {
		name   string
		entry  string
		update string
	}{
		{
			name:   "from account",
			entry:  "{AccountID: fromAccountID, Amount: cmd.TransferAmount, Direction: EntryDirectionDebit}",
			update: "adjustAccountBalance(ctx, tx, fromAccountID, cmd.TransferAmount, EntryDirectionDebit)",
		},
		{
			name:   "to account",
			entry:  "{AccountID: toAccountID, Amount: cmd.TransferAmount, Direction: EntryDirectionCredit}",
			update: "adjustAccountBalance(ctx, tx, toAccountID, cmd.TransferAmount, EntryDirectionCredit)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(source, tt.entry) {
				t.Fatalf("PostExternalTransfer does not record the expected %s ledger entry direction", tt.name)
			}
			if !strings.Contains(source, tt.update) {
				t.Fatalf("PostExternalTransfer records the expected %s entry direction but adjusts that account with a different direction", tt.name)
			}
		})
	}
}

package ledgerstress

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ledger-db/internal/ledgerstore"
	"ledger-db/internal/testdb"
)

func TestDeferredStressDrainsBeforeFinalValidation(t *testing.T) {
	ctx, _ := testdb.OpenIntegrationDB(t)
	percent := 80
	summary, err := Run(ctx, Config{
		DSN: os.Getenv("LEDGER_DB_DSN"), MigrationsDir: filepath.Join("..", "..", "db", "migrations"),
		Reset: true, Accounts: 10, Workers: 8, Operations: 200, MaxOpenConns: 8,
		ExternalTransferPercent: &percent, ReversalPercent: 10, DuplicatePercent: 20,
		DuplicateFanout: 3, ConflictPercent: 5,
		SettlementBalanceMode: ledgerstore.SettlementBalanceModeDeferredAsync,
		ValidationInterval:    10 * time.Millisecond, ProgressInterval: -1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !summary.Healthy || summary.Settlement == nil || summary.Settlement.Pending != 0 {
		t.Fatalf("unhealthy final result: %+v", summary)
	}
	if summary.Settlement.Applied == 0 || summary.Settlement.WorkerErrors != 0 {
		t.Fatalf("unexpected worker result: %+v", summary.Settlement)
	}
	if summary.Config.MinThinkTimeMS != 0 || summary.Config.MaxThinkTimeMS != 0 {
		t.Fatal("zero think time was replaced")
	}
}

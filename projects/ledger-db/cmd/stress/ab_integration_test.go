package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ledger-db/internal/ledgerschema"
	"ledger-db/internal/ledgerstress"
	"ledger-db/internal/testdb"
)

func TestABReturnsFailureForCompletedUnhealthyRuns(t *testing.T) {
	ctx, _ := testdb.OpenIntegrationDB(t)
	migrations := t.TempDir()
	for _, name := range ledgerschema.LocalSchemaMigrations {
		data, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "007_settlement_outbox.sql" {
			// The entries cannot account for this extra stored balance.
			data = append(data, []byte("\nupdate ledger_accounts set balance = 1 where name = 'Cash Settlement';\n")...)
		}
		if err := os.WriteFile(filepath.Join(migrations, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	code := runABTest(ctx, ledgerstress.Config{DSN: os.Getenv("LEDGER_DB_DSN"), MigrationsDir: migrations,
		Accounts: 2, Workers: 1, Operations: 2}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1; %s", code, stdout.String())
	}
	if strings.Count(stdout.String(), "FAILED") != 2 || strings.Contains(stdout.String(), "Deferred async minus sync:") {
		t.Fatalf("incorrect failure report: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "final validation found") {
		t.Fatalf("missing validation error: %s", stderr.String())
	}
}

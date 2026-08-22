package testdb

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"ledger-db/internal/ledgerschema"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultIntegrationTestDSN = "postgresql://ledger_db:password@localhost:5432/ledger_db"

func OpenIntegrationDB(t testing.TB) (context.Context, *sql.DB) {
	t.Helper()

	if os.Getenv("LEDGER_DB_INTEGRATION") != "1" {
		t.Skip("set LEDGER_DB_INTEGRATION=1 to run DB-backed ledger tests")
	}

	dsn := os.Getenv("LEDGER_DB_DSN")
	if dsn == "" {
		dsn = defaultIntegrationTestDSN
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	return ctx, db
}

func ResetIntegrationDB(t testing.TB, ctx context.Context, db *sql.DB) {
	t.Helper()

	err := ledgerschema.ApplyLocalSchema(ctx, db, filepath.Join(moduleRoot(t), "db", "migrations"))
	if err != nil {
		t.Fatal(err)
	}
}

func moduleRoot(t testing.TB) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate testdb source path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

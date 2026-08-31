package ledgerschema

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
)

var LocalSchemaMigrations = []string{
	"001_create_ledger_tables.sql",
	"002_create_external_transfers.sql",
	"003_seed_system_accounts.sql",
	"007_create_settlement_update_jobs.sql",
	"006_prevent_entry_mutations.sql",
}

func ApplyLocalSchema(ctx context.Context, db *sql.DB, migrationsDir string) error {
	for _, migration := range LocalSchemaMigrations {
		sqlText, err := os.ReadFile(filepath.Join(migrationsDir, migration))
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(sqlText)); err != nil {
			return err
		}
	}
	return nil
}

package migrate

import (
	"context"
	"errors"
	"fmt"

	goose "github.com/pressly/goose/v3"

	"weavelab.xyz/payer-sync/internal/db"
)

const migrationDir = "db/migrations"

func Run(ctx context.Context, command string) error {
	cfg := db.LoadConfigFromEnv()
	sqlDB, err := db.OpenSQLDB(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	switch command {
	case "up":
		return goose.Up(sqlDB, migrationDir)
	case "down":
		return goose.Down(sqlDB, migrationDir)
	case "status":
		return goose.Status(sqlDB, migrationDir)
	case "reset":
		return goose.Reset(sqlDB, migrationDir)
	default:
		return errors.New("unknown migration command; use one of: up, down, status, reset")
	}
}

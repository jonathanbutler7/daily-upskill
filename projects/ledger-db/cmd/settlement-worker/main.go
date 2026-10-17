package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ledger-db/internal/settlementworker"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dsn := os.Getenv("LEDGER_DB_DSN")
	if dsn == "" {
		dsn = "postgresql://ledger_db:password@localhost:5432/ledger_db"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	w := settlementworker.Worker{DB: db, DSN: dsn}
	if err := w.Run(ctx); err != nil {
		slog.Error("settlement worker", "error", err)
		os.Exit(1)
	}
}

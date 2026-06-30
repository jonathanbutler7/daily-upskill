package appops

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	seedersdk "github.com/jonathanbutler7/payer-sync-data-seeder/sdk"

	"weavelab.xyz/payer-sync/ingester"
	internaldb "weavelab.xyz/payer-sync/internal/db"
	"weavelab.xyz/payer-sync/internal/migrate"
	"weavelab.xyz/payer-sync/reconciler"
)

type SeedSummary struct {
	SeedRoundID string `json:"seed_round_id"`
	ERACount    int    `json:"era_count"`
	VCCCount    int    `json:"vcc_count"`
}

func OpenDB(ctx context.Context) (*pgxpool.Pool, error) {
	if err := migrate.Run(ctx, "up"); err != nil {
		return nil, fmt.Errorf("db migrate: %w", err)
	}
	pool, err := internaldb.OpenPool(ctx, internaldb.LoadConfigFromEnv())
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	return pool, nil
}

func Seed(ctx context.Context) (SeedSummary, error) {
	_ = godotenv.Load()

	client, err := seedersdk.NewClient(os.Getenv("SEEDER_BASE_URL"), os.Getenv("SEEDER_TOKEN"))
	if err != nil {
		return SeedSummary{}, fmt.Errorf("seed: create client: %w", err)
	}

	km, err := ingester.NewKeyManagerFromEnv()
	if err != nil {
		return SeedSummary{}, fmt.Errorf("seed: resolve private key source: %w", err)
	}
	if err := km.EnsureKeyPair(); err != nil {
		return SeedSummary{}, fmt.Errorf("seed: load key pair: %w", err)
	}
	if err := km.EnsureRegistered(ctx, client); err != nil {
		return SeedSummary{}, fmt.Errorf("seed: register public key: %w", err)
	}

	summary, err := client.CreateSeedBatch(ctx, nil)
	if err != nil {
		return SeedSummary{}, fmt.Errorf("seed: create batch: %w", err)
	}

	return SeedSummary{
		SeedRoundID: summary.SeedRoundID,
		ERACount:    int(summary.ERACount),
		VCCCount:    int(summary.VCCCount),
	}, nil
}

func Ingest(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("ingest: db pool is nil")
	}
	if err := ingester.Ingest(pool); err != nil {
		return fmt.Errorf("ingester: %w", err)
	}
	return nil
}

func Reconcile(ctx context.Context, pool *pgxpool.Pool) (reconciler.RunResult, error) {
	if pool == nil {
		return reconciler.RunResult{}, fmt.Errorf("reconcile: db pool is nil")
	}
	svc := reconciler.NewService(pool, reconciler.Config{})
	result, err := svc.Run(ctx)
	if err != nil {
		return result, fmt.Errorf("reconciler: %w", err)
	}
	return result, nil
}

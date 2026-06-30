package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"

	"weavelab.xyz/payer-sync/dashboard"
	"weavelab.xyz/payer-sync/internal/appops"
	internaldb "weavelab.xyz/payer-sync/internal/db"
	"weavelab.xyz/payer-sync/internal/migrate"
	"weavelab.xyz/payer-sync/processor"
)

const usage = `usage:
  go run . processor                        start the payment processor (long-running)
  go run . ingest                           fetch and store raw ERA/VCC files
  go run . reconcile                        match ERA and VCC records (run after ingest)
  go run . seed                             generate a fresh seed batch (dev only)
  go run . dashboard                        start the live demo dashboard
  go run . migrate <up|down|status|reset>   manage database migrations
`

func main() {
	configureLogging()
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "processor":
		err = runProcessor()
	case "ingest":
		err = runIngest()
	case "reconcile":
		err = runReconcile()
	case "seed":
		err = runSeed()
	case "dashboard":
		err = runDashboard()
	case "migrate":
		if len(os.Args) < 3 {
			fmt.Print(usage)
			os.Exit(2)
		}
		err = migrate.Run(context.Background(), os.Args[2])
	default:
		fmt.Print(usage)
		os.Exit(2)
	}

	if err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func configureLogging() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

// openDB runs migrations and returns an open connection pool.
// runSeed is a dev-only command that generates a fresh batch of ERA/VCC test data
// via the seeder service and registers the local public key if needed.
func runSeed() error {
	ctx := context.Background()

	summary, err := appops.Seed(ctx)
	if err != nil {
		return err
	}
	slog.Info("seed batch generated",
		"seed_round", summary.SeedRoundID,
		"era", summary.ERACount,
		"vcc", summary.VCCCount,
	)
	return nil
}

func runIngest() error {
	ctx := context.Background()

	pool, err := appops.OpenDB(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	return appops.Ingest(ctx, pool)
}

func runReconcile() error {
	ctx := context.Background()

	pool, err := appops.OpenDB(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	result, err := appops.Reconcile(ctx, pool)
	if err != nil {
		return fmt.Errorf("reconciler: %w", err)
	}
	slog.Info("reconciler run complete",
		"matched", result.MatchedCount,
		"expired_era", result.ExpiredERACount,
		"expired_vcc", result.ExpiredVCCCount,
		"run_id", result.RunID,
	)
	return nil
}

func runDashboard() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return dashboard.Run(ctx)
}

func runProcessor() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := appops.OpenDB(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Dedicated connection for LISTEN — must not be used for regular queries.
	cfg := internaldb.LoadConfigFromEnv()
	dsn, err := cfg.ConnectionString()
	if err != nil {
		return fmt.Errorf("db dsn: %w", err)
	}
	listenConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("listen conn: %w", err)
	}
	defer listenConn.Close(ctx)

	stripeAPIKey := os.Getenv("STRIPE_API_KEY")
	if stripeAPIKey == "" {
		return fmt.Errorf("STRIPE_API_KEY env var is required")
	}

	svc := processor.NewService(pool, processor.Config{
		Processor:  processor.NewStripeProcessor(processor.StripeConfig{APIKey: stripeAPIKey}),
		ListenConn: listenConn,
	})

	slog.Info("processor starting", "phase", "drain_then_listen")
	if err := svc.Run(ctx); err != nil {
		return fmt.Errorf("processor: %w", err)
	}
	slog.Info("processor stopped")
	return nil
}

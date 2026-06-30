package ingester

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	seedersdk "github.com/jonathanbutler7/payer-sync-data-seeder/sdk"

	"weavelab.xyz/payer-sync/internal/db/store"
	"weavelab.xyz/payer-sync/processor"
)

const (
	defaultBaseURL = "http://localhost:8080"
	defaultToken   = "local-dev-token"
)

// noopTrigger ignores reconciliation trigger calls. Replace with a real queue client in production.
type noopTrigger struct{}

func (n *noopTrigger) Trigger(_ context.Context, _ ReconcileTriggerRequest) error {
	return nil
}

// stripeTokenizer adapts processor.StripeProcessor to the ingester Tokenizer interface.
type stripeTokenizer struct {
	p *processor.StripeProcessor
}

type ingestionTarget struct {
	LocationID  string
	ExpectedNPI string
}

func (t *stripeTokenizer) CreatePaymentMethod(ctx context.Context, cardNumber, expMonth, expYear, cvv string) (string, error) {
	pm, err := t.p.CreatePaymentMethod(ctx, processor.CreatePaymentMethodRequest{
		CardNumber: cardNumber,
		ExpMonth:   expMonth,
		ExpYear:    expYear,
		CVV:        cvv,
	})
	if err != nil {
		return "", err
	}
	return pm.ID, nil
}

// Ingest wires all real dependencies and runs the ingester once.
// For local development this is called directly from main().
// In production, call this from a cron-triggered handler (e.g., a Cloud Scheduler job or k8s CronJob).
func Ingest(db store.DBTX) error {
	ctx := context.Background()

	baseURL := envOrDefault("SEEDER_BASE_URL", defaultBaseURL)
	token := envOrDefault("SEEDER_TOKEN", defaultToken)
	locationID := envOrDefault("LOCATION_ID", "location-001")
	ingestAllLocations := strings.EqualFold(envOrDefault("INGEST_ALL_LOCATIONS", "false"), "true")
	fingerprintKey := envOrDefault("CARD_FINGERPRINT_KEY", "")
	if fingerprintKey == "" {
		return fmt.Errorf("CARD_FINGERPRINT_KEY env var is required")
	}
	stripeAPIKey := envOrDefault("STRIPE_API_KEY", "")
	if stripeAPIKey == "" {
		return fmt.Errorf("STRIPE_API_KEY env var is required")
	}
	stripeCfg := processor.StripeConfig{
		APIKey: stripeAPIKey,
	}

	client, err := seedersdk.NewClient(baseURL, token)
	if err != nil {
		return fmt.Errorf("ingester: create client: %w", err)
	}

	km, err := NewKeyManagerFromEnv()
	if err != nil {
		return fmt.Errorf("ingester: resolve private key source: %w", err)
	}
	if err := km.EnsureKeyPair(); err != nil {
		return fmt.Errorf("ingester: ensure key pair: %w", err)
	}
	if err := km.EnsureRegistered(ctx, client); err != nil {
		return fmt.Errorf("ingester: register public key: %w", err)
	}

	dec := NewRSADecryptorFromKey(km.PrivateKey())
	rawStore := NewLocalRawFileStore(envOrDefault("RAW_STORAGE_DIR", "raw-storage"))

	tok := &stripeTokenizer{
		p: processor.NewStripeProcessor(stripeCfg),
	}
	q := store.New(db)

	locationFilter := ""
	if !ingestAllLocations {
		locationFilter = locationID
	}

	targets, err := listIngestionTargets(ctx, q, locationFilter)
	if err != nil {
		return fmt.Errorf("ingester: list ingestion targets: %w", err)
	}
	if len(targets) == 0 {
		if ingestAllLocations {
			return fmt.Errorf("ingester: no enabled ingestion targets found")
		}
		return fmt.Errorf("ingester: no enabled ingestion target found for location %q", locationID)
	}

	for _, target := range targets {
		if err := runIngestForLocation(
			ctx,
			db,
			client,
			dec,
			rawStore,
			tok,
			baseURL,
			target,
			fingerprintKey,
		); err != nil {
			return err
		}
	}
	return nil
}

func runIngestForLocation(
	ctx context.Context, db store.DBTX, client *seedersdk.Client,
	dec Decryptor, rawStore RawFileStore, tok Tokenizer, baseURL string,
	target ingestionTarget, fingerprintKey string,
) error {
	svc := NewService(
		Config{
			LocationID:     target.LocationID,
			ExpectedNPI:    target.ExpectedNPI,
			FingerprintKey: fingerprintKey,
		},
		client,
		dec,
		rawStore,
		NewRealIngesterStore(db),
		&noopTrigger{},
		tok,
	)

	start := time.Now()
	slog.Info("ingester run started",
		"server", baseURL,
		"location_id", target.LocationID,
	)

	result, err := svc.Run(ctx)
	if err != nil {
		return fmt.Errorf("ingester: run location %s: %w", target.LocationID, err)
	}

	elapsed := time.Since(start).Round(time.Millisecond)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			slog.Error("ingester file error", "file_key", e.Key, "error", e.Reason, "location_id", target.LocationID)
		}
		slog.Warn("ingester run complete",
			"server", baseURL,
			"location_id", target.LocationID,
			"total", result.Total,
			"processed", result.Processed,
			"duplicates", result.Duplicates,
			"file_errors", len(result.Errors),
			"duration", elapsed,
		)
	} else {
		slog.Info("ingester run complete",
			"server", baseURL,
			"location_id", target.LocationID,
			"total", result.Total,
			"processed", result.Processed,
			"duplicates", result.Duplicates,
			"file_errors", 0,
			"duration", elapsed,
		)
	}
	return nil
}

func listIngestionTargets(ctx context.Context, q *store.Queries, locationID string) ([]ingestionTarget, error) {
	if strings.TrimSpace(locationID) != "" {
		rows, err := q.ListEnabledIngestionTargetsForLocation(ctx, locationID)
		if err != nil {
			return nil, err
		}

		targets := make([]ingestionTarget, 0, len(rows))
		for _, row := range rows {
			if strings.TrimSpace(row.LocationID) == "" || strings.TrimSpace(row.ProviderNpi) == "" {
				continue
			}
			targets = append(targets, ingestionTarget{
				LocationID:  row.LocationID,
				ExpectedNPI: row.ProviderNpi,
			})
		}
		return targets, nil
	}

	rows, err := q.ListEnabledIngestionTargets(ctx)
	if err != nil {
		return nil, err
	}

	targets := make([]ingestionTarget, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.LocationID) == "" || strings.TrimSpace(row.ProviderNpi) == "" {
			continue
		}
		targets = append(targets, ingestionTarget{
			LocationID:  row.LocationID,
			ExpectedNPI: row.ProviderNpi,
		})
	}

	return targets, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

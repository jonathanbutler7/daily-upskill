package ledgerstress

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"

	"ledger-db/internal/ledgerstore"
	"ledger-db/internal/ledgervalidator"
)

func TestFakeStressAccountUsesReadableMerchantData(t *testing.T) {
	profile := fakeStressAccount(1)

	if profile.Name != "Austin Market Co" {
		t.Fatalf("unexpected account name: %q", profile.Name)
	}
	if profile.Description != "Merchant wallet for Austin, TX" {
		t.Fatalf("unexpected account description: %q", profile.Description)
	}
	if strings.Contains(profile.Name, "Stress Account") {
		t.Fatalf("account name should not expose stress scaffolding: %q", profile.Name)
	}
}

func TestFakeStressAccountCyclesWithLocationSuffix(t *testing.T) {
	profile := fakeStressAccount(13)

	if profile.Name != "Austin Market Co - Location 2" {
		t.Fatalf("unexpected cycled account name: %q", profile.Name)
	}
}

func TestFakeReferenceLabelsStayShortAndUnique(t *testing.T) {
	runLabel := runLabelFromID("c65e7a50-8c44-4282-b318-4e2d9b6127af")

	if runLabel != "c65e7a50" {
		t.Fatalf("unexpected run label: %q", runLabel)
	}

	seedReference := fakeExternalReference(runLabel, "seed", 3, 0)
	if seedReference != "ach-settlement-c65e7a50-0003" {
		t.Fatalf("unexpected seed reference: %q", seedReference)
	}

	operationKey := fakeIdempotencyKey(runLabel, "deposit", 2, 17, 44)
	if operationKey != "demo-c65e7a50-deposit-w02-op000017-seq000044" {
		t.Fatalf("unexpected operation idempotency key: %q", operationKey)
	}
}

func TestWriteTextEventRendersDashboard(t *testing.T) {
	var output bytes.Buffer

	writeTextEvent(&output, progressEvent{
		Type:   "progress",
		At:     timeNowForTest(),
		RunID:  "c65e7a50-8c44-4282-b318-4e2d9b6127af",
		Target: 100,
		Config: SummaryConfig{
			Accounts:           5,
			Workers:            2,
			Operations:         100,
			MaxOpenConns:       6,
			OperationTimeoutMS: 30000,
		},
		Stats: StatsSnapshot{
			Total:         25,
			Success:       24,
			Failure:       1,
			OpsPerSecond:  12.5,
			ByOperation:   map[string]int64{"deposit": 4, "transfer": 18, "withdrawal": 3},
			ByErrorCode:   map[string]int64{"internal.unknown": 1},
			ErrorSamples:  []ErrorSample{{Code: "internal.unknown", Category: "internal", Operation: "transfer", Message: "context deadline exceeded while waiting for a connection from the pool", Count: 1}},
			RetryAttempts: 2,
			LatencyMS:     LatencySnapshot{Avg: 3.1, P50: 2.9, P95: 8.4, P99: 10.1, Max: 12.0},
		},
	})

	rendered := output.String()
	for _, expected := range []string{
		"LedgerDB Status",
		"Workload",
		"Latency",
		"Operations",
		"DB Pool",
		"Validation",
		"Errors",
		"internal.unknown",
		"msg=context deadline exceeded while waiting for a connection from the pool",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("dashboard output missing %q:\n%s", expected, rendered)
		}
	}
	if strings.Contains(rendered, "progress ops=") {
		t.Fatalf("dashboard should not use the old one-line progress format:\n%s", rendered)
	}
}

func TestWriteTextEventRendersValidationExpectedAndActualCounts(t *testing.T) {
	var output bytes.Buffer

	writeTextEvent(&output, finalEvent{
		At: timeNowForTest(),
		Summary: Summary{
			RunID: "c65e7a50-8c44-4282-b318-4e2d9b6127af",
			Config: SummaryConfig{
				Reset:              true,
				Accounts:           5,
				Workers:            2,
				Operations:         50,
				MaxOpenConns:       6,
				OperationTimeoutMS: 30000,
			},
			Stats: StatsSnapshot{
				Total:                 50,
				Success:               48,
				Failure:               2,
				ByOperation:           map[string]int64{"deposit": 8, "transfer": 34, "withdrawal": 8},
				BySuccessfulOperation: map[string]int64{"deposit": 7, "transfer": 34, "withdrawal": 7},
				LatencyMS:             LatencySnapshot{Avg: 3.1, P50: 2.9, P95: 8.4, P99: 10.1, Max: 12.0},
			},
			DBStats: DBStats{OpenConnections: 2, Idle: 2, MaxOpenConnections: 6},
			Performance: PerformanceGrade{
				Grade:           "B",
				Score:           82,
				Summary:         "good throughput; tail latency needs attention",
				ThroughputRPS:   725.2,
				P95MS:           81.6,
				P99MS:           256.0,
				PoolWaitPerOpMS: 2.9,
				MaxLockWaiters:  1,
			},
			FinalValidation: ledgervalidator.Result{
				Healthy: true,
				Summary: ledgervalidator.Summary{
					AccountCount:          6,
					TransactionCount:      53,
					EntryCount:            106,
					ExternalTransferCount: 19,
					ReversalCount:         0,
					IssueCount:            0,
				},
			},
			Healthy: true,
		},
	})

	rendered := output.String()
	for _, expected := range []string{
		"Check        Expected     Actual Status",
		"Accounts            6          6",
		"Txns               53         53",
		"Entries           106        106",
		"External           19         19",
		"Reversals           0          0",
		"Issues              0          0",
		"grade:",
		"B/82",
		"pool_wait/op=2.90ms",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("dashboard output missing %q:\n%s", expected, rendered)
		}
	}
}

func TestRandomAccountPairCanRouteToHotAccounts(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	accounts := []ledgerstore.AccountID{10, 11, 12, 13, 14}
	config := Config{HotAccounts: 2, HotTransferPercent: 100}

	for i := 0; i < 20; i++ {
		from, to := randomAccountPair(random, accounts, config)
		if from > 11 || to > 11 {
			t.Fatalf("expected hot-account pair, got from=%d to=%d", from, to)
		}
		if from == to {
			t.Fatalf("expected distinct accounts, got %d", from)
		}
	}
}

func TestSettlementAccountForIndexCyclesBuckets(t *testing.T) {
	buckets := []ledgerstore.AccountID{101, 102, 103}

	if got := settlementAccountForIndex(buckets, 0); got != 101 {
		t.Fatalf("unexpected bucket for index 0: %d", got)
	}
	if got := settlementAccountForIndex(buckets, 4); got != 102 {
		t.Fatalf("unexpected bucket for index 4: %d", got)
	}
	if got := settlementAccountForIndex(nil, 4); got != 0 {
		t.Fatalf("expected default settlement account marker, got %d", got)
	}
}

func timeNowForTest() time.Time {
	return time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
}

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
			Total:                25,
			Success:              24,
			Failure:              1,
			RequestAttempts:      30,
			DuplicateRequests:    5,
			ConcurrentDuplicates: 2,
			IdempotentReplays:    5,
			PoolTimeouts:         1,
			OpsPerSecond:         12.5,
			ByOperation:          map[string]int64{"deposit": 4, "transfer": 18, "withdrawal": 3},
			ByErrorCode:          map[string]int64{"db.unavailable": 1},
			ErrorSamples:         []ErrorSample{{Code: "db.unavailable", Category: "db", Operation: "transfer", Message: "context deadline exceeded while waiting for a connection from the pool", Count: 1}},
			RetryAttempts:        2,
			LatencyMS:            LatencySnapshot{Avg: 3.1, P50: 2.9, P95: 8.4, P99: 10.1, Max: 12.0},
		},
	})

	rendered := output.String()
	for _, expected := range []string{
		"LedgerDB Status",
		"Workload",
		"Goroutines",
		"Workers  2 worker goroutines",
		"Runtime",
		"goroutines now",
		"Other",
		"non-worker goroutines",
		"Operations",
		"DB Pool",
		"Idempotency",
		"Concurrency",
		"Requests  30 total for 25 logical ops",
		"Replays  5 same-transaction returns",
		"Spread     accounts/worker=2.5 hot=off",
		"Failed  pool_timeouts=1 server_rejects=0 lock=0",
		"Validation",
		"Errors",
		"db.unavailable",
		"msg=context deadline exceeded while waiting for a connection from the pool",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("dashboard output missing %q:\n%s", expected, rendered)
		}
	}
	if strings.Contains(rendered, "progress ops=") {
		t.Fatalf("dashboard should not use the old one-line progress format:\n%s", rendered)
	}
	if strings.Contains(rendered, "Latency (bar=max)") {
		t.Fatalf("dashboard should not render the latency panel:\n%s", rendered)
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
				Total:                   50,
				Success:                 48,
				Failure:                 2,
				RequestAttempts:         55,
				DuplicateRequests:       5,
				ConcurrentDuplicates:    2,
				IdempotentReplays:       5,
				IdempotencyConflicts:    1,
				PoolTimeouts:            1,
				ServerConnectionRejects: 1,
				ByOperation:             map[string]int64{"deposit": 8, "transfer": 34, "withdrawal": 8},
				BySuccessfulOperation:   map[string]int64{"deposit": 7, "transfer": 34, "withdrawal": 7},
				LatencyMS:               LatencySnapshot{Avg: 3.1, P50: 2.9, P95: 8.4, P99: 10.1, Max: 12.0},
			},
			DBStats: DBStats{OpenConnections: 2, Idle: 2, MaxOpenConnections: 6},
			Performance: PerformanceGrade{
				Grade:           "B",
				Score:           82,
				Summary:         "good throughput; pool queueing is visible",
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
		"worker_goroutines=2",
		"runtime_goroutines=",
		"pool_wait/op=2.90ms",
		"pool_timeouts=1 server_rejects=1",
		"idem:",
		"requests=55 duplicate_extra=5 replays=5 conflicts=1 unexpected=0",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("dashboard output missing %q:\n%s", expected, rendered)
		}
	}
}

func TestWriteTextEventRendersValidationIssueDetails(t *testing.T) {
	var output bytes.Buffer
	accountID := int64(17)
	transactionID := int64(42)
	storedAmount := int64(100)
	derivedAmount := int64(50)
	delta := int64(50)

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
				Success:               50,
				RequestAttempts:       50,
				ByOperation:           map[string]int64{"deposit": 8, "transfer": 34, "withdrawal": 8},
				BySuccessfulOperation: map[string]int64{"deposit": 8, "transfer": 34, "withdrawal": 8},
			},
			FinalValidation: ledgervalidator.Result{
				Healthy: false,
				Summary: ledgervalidator.Summary{
					AccountCount:          6,
					TransactionCount:      55,
					EntryCount:            110,
					ExternalTransferCount: 21,
					ReversalCount:         0,
					IssueCount:            3,
				},
				Issues: []ledgervalidator.Issue{
					{
						Check:         ledgervalidator.CheckAccountBalances,
						Code:          "account_balance_mismatch",
						Message:       "account 17 stored balance 100 does not match derived balance 50",
						AccountID:     &accountID,
						StoredAmount:  &storedAmount,
						DerivedAmount: &derivedAmount,
						Delta:         &delta,
					},
					{
						Check:         ledgervalidator.CheckTransactionShape,
						Code:          "transaction_entry_shape_mismatch",
						Message:       "posted transaction 42 has entries that do not match its from/to amount",
						TransactionID: &transactionID,
					},
					{
						Check:         ledgervalidator.CheckTransactionShape,
						Code:          "transaction_entry_shape_mismatch",
						Message:       "posted transaction 42 has entries that do not match its from/to amount",
						TransactionID: &transactionID,
					},
				},
			},
			Healthy: false,
		},
	})

	rendered := output.String()
	for _, expected := range []string{
		"Final audit failed",
		"Issues              0          3",
		"Issue detail showing 2 of 2 check/code group(s)",
		"transaction_shape transaction_entry_shape_mismatch count=2 txn=42",
		"account_balances account_balance_mismatch count=1 account=17 stored=100 derived=50 delta=50",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("dashboard output missing %q:\n%s", expected, rendered)
		}
	}
}

func TestSummarizeDuplicateResultsRecordsIdempotentReplays(t *testing.T) {
	collector := newCollector(timeNowForTest())

	err := summarizeDuplicateResults("transfer", []requestResult{
		{transactionID: 42},
		{transactionID: 42},
		{transactionID: 42},
	}, collector)
	if err != nil {
		t.Fatalf("summarizeDuplicateResults returned error: %v", err)
	}

	snapshot := collector.snapshot()
	if snapshot.IdempotentReplays != 2 {
		t.Fatalf("idempotent replays = %d, want 2", snapshot.IdempotentReplays)
	}
	if snapshot.IdempotencyUnexpected != 0 {
		t.Fatalf("unexpected idempotency failures = %d, want 0", snapshot.IdempotencyUnexpected)
	}
}

func TestSummarizeDuplicateResultsRejectsDifferentTransactionIDs(t *testing.T) {
	collector := newCollector(timeNowForTest())

	err := summarizeDuplicateResults("transfer", []requestResult{
		{transactionID: 42},
		{transactionID: 43},
	}, collector)
	if err == nil {
		t.Fatalf("expected duplicate transaction id mismatch error")
	}

	snapshot := collector.snapshot()
	if snapshot.IdempotencyUnexpected != 1 {
		t.Fatalf("unexpected idempotency failures = %d, want 1", snapshot.IdempotencyUnexpected)
	}
	if !collector.hasUnexpectedErrors() {
		t.Fatalf("collector should treat idempotency mismatch as unexpected")
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

func timeNowForTest() time.Time {
	return time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
}

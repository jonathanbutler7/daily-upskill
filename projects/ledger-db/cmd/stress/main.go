package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ledger-db/internal/ledgerstore"
	"ledger-db/internal/ledgerstress"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type cliOptions struct {
	config                  ledgerstress.Config
	externalTransferPercent int
	help                    bool
	fullHelp                bool
	abTest                  bool
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	options := cliOptions{
		config: ledgerstress.Config{
			DSN:             os.Getenv("LEDGER_DB_DSN"),
			Reset:           true,
			ReversalPercent: 10,
		},
	}
	flags := newFlagSet(&options, stdout, stderr)

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if options.help {
		writeHelp(stdout)
		return 0
	}
	if options.fullHelp {
		writeFullHelp(stdout, flags)
		return 0
	}

	options.config.ExternalTransferPercent = &options.externalTransferPercent
	options.config.Output = stdout

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if options.abTest {
		return runABTest(ctx, options.config, stdout, stderr)
	}

	if _, err := ledgerstress.Run(ctx, options.config); err != nil {
		fmt.Fprintf(stderr, "stress run failed: %v\n", err)
		return 1
	}

	return 0
}

func newFlagSet(options *cliOptions, stdout io.Writer, stderr io.Writer) *flag.FlagSet {
	config := &options.config
	flags := flag.NewFlagSet("stress", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		writeHelp(stdout)
	}

	flags.BoolVar(&options.help, "help", false, "show this help")
	flags.BoolVar(&options.help, "h", false, "show this help")
	flags.BoolVar(&options.fullHelp, "help-full", false, "show all flags and longer examples")
	flags.BoolVar(&options.abTest, "ab-test", false, "run sync and deferred async modes and compare contention")
	flags.IntVar(&config.Accounts, "accounts", 25, "number of stress wallet accounts to create")
	flags.IntVar(&config.Workers, "workers", 10, "number of concurrent workers")
	flags.IntVar(&config.Operations, "operations", 1000, "number of logical ledger operations to attempt")
	flags.IntVar(&config.MaxRetries, "max-retries", 3, "max retries for retryable database errors")
	flags.IntVar(&config.MaxOpenConns, "max-open-conns", 0, "database max open connections; 0 means min(workers + 4, 50)")
	flags.IntVar(&options.externalTransferPercent, "external-transfer-percent", 30, "percentage of operations that are external deposits or withdrawals; split evenly")
	flags.IntVar(&config.HotAccounts, "hot-accounts", 0, "number of seeded accounts to treat as hot transfer accounts; 0 disables hot-account mode")
	flags.IntVar(&config.HotTransferPercent, "hot-transfer-percent", 0, "percentage of wallet transfers routed within the hot-account set")
	flags.IntVar(&config.DuplicatePercent, "duplicate-percent", 0, "percentage of logical operations that fan out concurrent duplicate requests")
	flags.IntVar(&config.DuplicateFanout, "duplicate-fanout", 2, "number of same-key concurrent requests for duplicate operations")
	flags.IntVar(&config.ConflictPercent, "conflict-percent", 0, "percentage of successful operations followed by an expected idempotency conflict")
	flags.IntVar(&config.ReversalPercent, "reversal-percent", 10, "percentage of logical operations that attempt to reverse a prior transaction")
	flags.DurationVar(&config.MinThinkTime, "think-min", 25*time.Millisecond, "minimum random pause between worker operations")
	flags.DurationVar(&config.MaxThinkTime, "think-max", 250*time.Millisecond, "maximum random pause between worker operations")
	flags.DurationVar(&config.ProgressInterval, "progress-interval", time.Second, "progress interval; set a negative duration like -1s to disable")

	return flags
}

func writeHelp(output io.Writer) {
	fmt.Fprintln(output, "LedgerDB stress test")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Usage: go run ./cmd/stress [flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Common runs:")
	fmt.Fprintln(output, "  Smoke")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=5 -workers=2 -operations=50")
	fmt.Fprintln(output, "  Duplicate/idempotency")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=25 -workers=20 -operations=1000 -duplicate-percent=10 -duplicate-fanout=3 -conflict-percent=2")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Most-used flags:")
	fmt.Fprintln(output, "  -accounts, -workers, -operations, -max-open-conns")
	fmt.Fprintln(output, "  -duplicate-percent, -duplicate-fanout, -conflict-percent")
	fmt.Fprintln(output, "  -reversal-percent, -external-transfer-percent")
	fmt.Fprintln(output, "  -ab-test")
	fmt.Fprintln(output, "  -hot-accounts, -hot-transfer-percent")
	fmt.Fprintln(output, "  -think-min, -think-max, -progress-interval")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Scoring:")
	fmt.Fprintln(output, "  Score starts at 100. Grades: A>=90, B>=80, C>=70, D>=60, F<60.")
	fmt.Fprintln(output, "  Penalties come from correctness, failures, retries, P95/P99 latency,")
	fmt.Fprintln(output, "  pool wait per op, low RPS, and lock waiters.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "More detail:")
	fmt.Fprintln(output, "  go run ./cmd/stress --help-full")
	fmt.Fprintln(output, "  --help-full includes sample commands for targeting A/B/C/D/F scores.")
	fmt.Fprintln(output, "  docs/stress-test.md")
}

func writeFullHelp(output io.Writer, flags *flag.FlagSet) {
	fmt.Fprintln(output, "LedgerDB stress test")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Usage: go run ./cmd/stress [flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Good starting point:")
	fmt.Fprintln(output, "  Small smoke test")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=5 -workers=2 -operations=50")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Heavier comparison runs:")
	fmt.Fprintln(output, "  Baseline local load")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=150 -operations=50000 -max-open-conns=75")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "  Hot-account skew check")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=150 -operations=50000 -hot-accounts=5 -hot-transfer-percent=20")
	fmt.Fprintln(output, "    This intentionally creates row-lock contention and can run much longer.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "  Duplicate request and idempotency check")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=25 -workers=20 -operations=1000 -duplicate-percent=10 -duplicate-fanout=3 -conflict-percent=2")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Core workload flags:")
	printFlagGroup(output, flags, []string{
		"accounts",
		"workers",
		"operations",
		"external-transfer-percent",
		"think-min",
		"think-max",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Database flags:")
	printFlagGroup(output, flags, []string{
		"max-open-conns",
		"max-retries",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Stress-shape flags:")
	printFlagGroup(output, flags, []string{
		"hot-accounts",
		"hot-transfer-percent",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Concurrent request and idempotency flags:")
	fmt.Fprintln(output, "  -operations stays the number of logical ledger operations.")
	fmt.Fprintln(output, "  Healthy duplicate runs show replay hits, expected conflicts, and unexpected=0.")
	printFlagGroup(output, flags, []string{
		"duplicate-percent",
		"duplicate-fanout",
		"conflict-percent",
		"reversal-percent",
		"ab-test",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Output and validation flags:")
	printFlagGroup(output, flags, []string{
		"progress-interval",
		"help",
		"h",
		"help-full",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "How to read the results:")
	fmt.Fprintln(output, "  RPS shows completed operations per second.")
	fmt.Fprintln(output, "  Goroutines shows configured worker goroutines and current runtime goroutines.")
	fmt.Fprintln(output, "  DB Pool and Lock waiters show connection and row-lock pressure.")
	fmt.Fprintln(output, "  See docs/stress-test.md for output fields and interpretation.")
	fmt.Fprintln(output, "  Validation compares expected and actual ledger counts at the end.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Current scoring system:")
	fmt.Fprintln(output, "  Score starts at 100. Letter grades are A>=90, B>=80, C>=70, D>=60, F<60.")
	fmt.Fprintln(output, "  Correctness failed: -40")
	fmt.Fprintln(output, "  Any operation failure: -15")
	fmt.Fprintln(output, "  Any retry: -10")
	fmt.Fprintln(output, "  P99 latency: >=100ms -8, >=250ms -15, >=500ms -20, >=1000ms -25")
	fmt.Fprintln(output, "  P95 latency: >=50ms -5, >=100ms -10, >=250ms -15")
	fmt.Fprintln(output, "  Pool wait per op: >=1ms -5, >=5ms -10, >=10ms -15")
	fmt.Fprintln(output, "  Throughput: <500 RPS -5, <100 RPS -10")
	fmt.Fprintln(output, "  Lock waiters: -2 each, capped at -10")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "How to move scores:")
	fmt.Fprintln(output, "  A: keep failures/retries at 0, P95 under 50ms, P99 under 100ms,")
	fmt.Fprintln(output, "     pool_wait/op under 1ms, RPS at or above 500, and lock_waiters_max=0.")
	fmt.Fprintln(output, "  B/C/D: raise workers, lower -max-open-conns, reduce think time, or add")
	fmt.Fprintln(output, "     hot-account skew until latency, pool wait, or lock waiters cross thresholds.")
	fmt.Fprintln(output, "  F: combine correctness failure, operation failures, very high tail latency,")
	fmt.Fprintln(output, "     pool queueing, low RPS, or lock waiters until score drops below 60.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Sample grade targets:")
	fmt.Fprintln(output, "  Local results vary by machine and Postgres settings. Use these as starting")
	fmt.Fprintln(output, "  points, then compare the final P95/P99, pool_wait/op, RPS, and lock_waiters.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "  Target A")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=25 -operations=10000 -max-open-conns=50")
	fmt.Fprintln(output, "  Target B")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=75 -operations=20000 -max-open-conns=25")
	fmt.Fprintln(output, "  Target C")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=150 -operations=50000 -max-open-conns=75")
	fmt.Fprintln(output, "  Target D")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=150 -operations=50000 -max-open-conns=25")
	fmt.Fprintln(output, "  Target F")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=200 -operations=50000 -max-open-conns=10 -think-min=0s -think-max=0s -hot-accounts=2 -hot-transfer-percent=80")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "  If a target scores too high, lower -max-open-conns, reduce think time, or")
	fmt.Fprintln(output, "  raise -hot-transfer-percent. If it scores too low, do the opposite.")
}

type abResult struct {
	label   string
	summary ledgerstress.Summary
	err     error
}

func runABTest(ctx context.Context, config ledgerstress.Config, stdout, stderr io.Writer) int {
	config.Output = io.Discard
	// Keep the progress reporter running even though its output is discarded.
	// It samples PostgreSQL lock waiters during the workload and records the peak.
	if config.ProgressInterval < 0 {
		config.ProgressInterval = time.Second
	}
	config.ValidationInterval = 0
	config.ReversalPercent = 0

	fmt.Fprintln(stdout, "LedgerDB contention A/B test")
	fmt.Fprintln(stdout, "Both runs use the same workload settings and reset the database before running.")
	fmt.Fprintln(stdout, "Reversals are disabled so this measures external-transfer settlement behavior only.")
	fmt.Fprintln(stdout)

	results := []abResult{
		{label: "sync", summary: runABMode(ctx, config, ledgerstore.SettlementBalanceModeSync, stderr)},
		{label: "deferred async", summary: runABMode(ctx, config, ledgerstore.SettlementBalanceModeDeferredAsync, stderr)},
	}

	fmt.Fprintln(stdout, "Results (lower is better for latency, pool wait, and lock waiters):")
	fmt.Fprintf(stdout, "%-22s %8s %12s %12s %12s %16s %16s %18s\n", "Mode", "Grade", "RPS", "P95 ms", "P99 ms", "Pool wait/op", "Lock waiters", "Validation")
	fmt.Fprintf(stdout, "%-22s %8s %12s %12s %12s %16s %16s %18s\n", "----------------------", "--------", "------------", "------------", "------------", "----------------", "----------------", "------------------")

	for i := range results {
		result := &results[i]
		grade := "-"
		if result.summary.Performance.Grade != "" {
			grade = fmt.Sprintf("%s/%d", result.summary.Performance.Grade, result.summary.Performance.Score)
		}
		validation := "healthy"
		if !result.summary.Healthy {
			validation = "expected lag/fail"
		}
		fmt.Fprintf(stdout, "%-22s %8s %12.1f %12.2f %12.2f %16.2f %16d %18s\n",
			result.label,
			grade,
			result.summary.Stats.OpsPerSecond,
			result.summary.Stats.LatencyMS.P95,
			result.summary.Stats.LatencyMS.P99,
			result.summary.Performance.PoolWaitPerOpMS,
			result.summary.Stats.MaxLockWaiters,
			validation,
		)
	}

	if len(results) == 2 && results[0].summary.RunID != "" && results[1].summary.RunID != "" {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Deferred async minus sync:")
		fmt.Fprintf(stdout, "  RPS:           %+0.1f/sec\n", results[1].summary.Stats.OpsPerSecond-results[0].summary.Stats.OpsPerSecond)
		fmt.Fprintf(stdout, "  P95 latency:   %+0.2f ms\n", results[1].summary.Stats.LatencyMS.P95-results[0].summary.Stats.LatencyMS.P95)
		fmt.Fprintf(stdout, "  P99 latency:   %+0.2f ms\n", results[1].summary.Stats.LatencyMS.P99-results[0].summary.Stats.LatencyMS.P99)
		fmt.Fprintf(stdout, "  Pool wait/op:  %+0.2f ms\n", results[1].summary.Performance.PoolWaitPerOpMS-results[0].summary.Performance.PoolWaitPerOpMS)
		fmt.Fprintf(stdout, "  Lock waiters:  %+d\n", results[1].summary.Stats.MaxLockWaiters-results[0].summary.Stats.MaxLockWaiters)
	}

	for _, result := range results {
		if result.summary.RunID == "" {
			return 1
		}
	}
	return 0
}

func runABMode(ctx context.Context, base ledgerstress.Config, mode ledgerstore.SettlementBalanceMode, stderr io.Writer) ledgerstress.Summary {
	runConfig := base
	runConfig.SettlementBalanceMode = mode
	summary, err := ledgerstress.Run(ctx, runConfig)
	if err != nil && summary.RunID == "" {
		fmt.Fprintf(stderr, "%s run failed before producing metrics: %v\n", mode, err)
	}
	return summary
}

func printFlagGroup(output io.Writer, flags *flag.FlagSet, names []string) {
	for _, name := range names {
		flagValue := flags.Lookup(name)
		if flagValue == nil {
			continue
		}
		fmt.Fprintf(output, "  -%-22s default=%-18s %s\n", flagValue.Name, flagValue.DefValue, flagValue.Usage)
	}
}

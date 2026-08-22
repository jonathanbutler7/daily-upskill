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

	"ledger-db/internal/ledgerstress"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type cliOptions struct {
	config ledgerstress.Config
	help   bool
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	var options cliOptions
	flags := newFlagSet(&options, stdout, stderr)

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if options.help {
		writeHelp(stdout, flags)
		return 0
	}

	options.config.Output = stdout

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
		writeHelp(stdout, flags)
	}

	flags.BoolVar(&options.help, "help", false, "show this help")
	flags.BoolVar(&options.help, "h", false, "show this help")
	flags.StringVar(&config.DSN, "dsn", envOrDefault("LEDGER_DB_DSN", ledgerstress.DefaultDSN), "Postgres DSN")
	flags.StringVar(&config.MigrationsDir, "migrations-dir", "db/migrations", "directory containing local schema migrations")
	flags.BoolVar(&config.Reset, "reset", true, "reset the local schema before running")
	flags.IntVar(&config.Accounts, "accounts", 25, "number of stress wallet accounts to create")
	flags.IntVar(&config.Workers, "workers", 10, "number of concurrent workers")
	flags.IntVar(&config.Operations, "operations", 1000, "number of operations to attempt")
	flags.Int64Var(&config.SeedBalance, "seed-balance", 100_000, "starting balance for each stress account")
	flags.Int64Var(&config.MaxAmount, "max-amount", 100, "maximum amount per operation")
	flags.IntVar(&config.MaxRetries, "max-retries", 3, "max retries for retryable database errors")
	flags.IntVar(&config.MaxOpenConns, "max-open-conns", 0, "database max open connections; 0 means min(workers + 4, 50)")
	flags.IntVar(&config.SettlementBuckets, "settlement-buckets", 1, "number of ACH settlement bucket accounts; 1 uses the default Cash Settlement account")
	flags.IntVar(&config.HotAccounts, "hot-accounts", 0, "number of seeded accounts to treat as hot transfer accounts; 0 disables hot-account mode")
	flags.IntVar(&config.HotTransferPercent, "hot-transfer-percent", 0, "percentage of wallet transfers routed within the hot-account set")
	flags.DurationVar(&config.OperationTimeout, "operation-timeout", 30*time.Second, "timeout for each ledger operation")
	flags.DurationVar(&config.MinThinkTime, "think-min", 25*time.Millisecond, "minimum random pause between worker operations")
	flags.DurationVar(&config.MaxThinkTime, "think-max", 250*time.Millisecond, "maximum random pause between worker operations")
	flags.DurationVar(&config.ProgressInterval, "progress-interval", time.Second, "progress interval; set a negative duration like -1s to disable")
	flags.DurationVar(&config.ValidationInterval, "validation-interval", 0, "optional in-run validation interval; disabled by default")
	flags.StringVar(&config.OutputFormat, "format", ledgerstress.OutputFormatText, "output format: text or json")

	return flags
}

func writeHelp(output io.Writer, flags *flag.FlagSet) {
	fmt.Fprintln(output, "LedgerDB stress test")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  go run ./cmd/stress [flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "What it does:")
	fmt.Fprintln(output, "  Resets the local ledger schema by default, creates fake merchant")
	fmt.Fprintln(output, "  wallet accounts, seeds balances, runs concurrent deposits, withdrawals,")
	fmt.Fprintln(output, "  and wallet transfers, then finishes with a validator audit and")
	fmt.Fprintln(output, "  performance grade.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Good starting point:")
	fmt.Fprintln(output, "  Small smoke test")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=5 -workers=2 -operations=50")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Heavier comparison runs:")
	fmt.Fprintln(output, "  Baseline local load")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=150 -operations=50000 -max-open-conns=75")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "  Settlement contention check")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=150 -operations=50000 -settlement-buckets=16")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "  Hot-account skew check")
	fmt.Fprintln(output, "    go run ./cmd/stress -accounts=500 -workers=150 -operations=50000 -hot-accounts=5 -hot-transfer-percent=20")
	fmt.Fprintln(output, "    This intentionally creates row-lock contention and can run much longer.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "  JSON stream for scripts")
	fmt.Fprintln(output, "    go run ./cmd/stress -format=json")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Core workload flags:")
	printFlagGroup(output, flags, []string{
		"accounts",
		"workers",
		"operations",
		"seed-balance",
		"max-amount",
		"think-min",
		"think-max",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Database and reset flags:")
	printFlagGroup(output, flags, []string{
		"dsn",
		"migrations-dir",
		"reset",
		"max-open-conns",
		"operation-timeout",
		"max-retries",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Stress-shape flags:")
	printFlagGroup(output, flags, []string{
		"settlement-buckets",
		"hot-accounts",
		"hot-transfer-percent",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Output and validation flags:")
	printFlagGroup(output, flags, []string{
		"progress-interval",
		"validation-interval",
		"format",
		"help",
		"h",
	})
	fmt.Fprintln(output)
	fmt.Fprintln(output, "How to read the results:")
	fmt.Fprintln(output, "  RPS shows completed operations per second.")
	fmt.Fprintln(output, "  P95/P99 show tail latency; these matter more than the average.")
	fmt.Fprintln(output, "  DB Pool queue wait means workers were waiting for a connection.")
	fmt.Fprintln(output, "  Lock waiters means Postgres sessions were waiting on row locks.")
	fmt.Fprintln(output, "  Validation compares expected and actual ledger counts at the end.")
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

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

# Stress Test Plan

The first stress-test runner is a local Go CLI:

```bash
go run ./cmd/stress
```

Use `go run ./cmd/stress --help` to see the run types, common examples, and all
available flags grouped by purpose.

It resets the local ledger schema by default, creates stress accounts, seeds each
account through posted external deposits, runs concurrent workers, prints live
progress, and finishes with a read-only validator audit.

## Runner Shape

The runner is split into two parts:

- `internal/ledgerstress`: reusable stress harness
- `cmd/stress`: CLI flag parsing and process exit behavior

This keeps stress-test code out of the write path and out of the validator.

## Current Runner

The current runner calls the Go command boundary directly. It exercises:

- posted external deposits
- posted external withdrawals
- wallet-to-wallet transfers
- idempotency keys for every attempted operation
- retry behavior for retryable database errors
- randomized operation mix, account choice, amount, and worker pacing
- final committed-state validation through `internal/ledgervalidator`

This runner is for write-path and database contention. It does not measure HTTP
server overhead.

## Visibility

The runner prints a live terminal dashboard by default:

```text
LedgerDB Status  Health RUNNING  Run d17edc05  18:10:42  accounts=5 workers=2 ops=50
Local Postgres . USD ledger . randomized deposits, withdrawals, and wallet transfers

Workload                                          Latency
Ops      [###########-----------------]  40.0%    Avg      [------------------]  3.89 ms
OK       [###########-----------------]  20       P95      [------------------]  6.10 ms
Failed   [----------------------------]  0        P99      [------------------] 10.30 ms

Operations                                        DB Pool
Deposit    [###---------------------] 6           Held    [#######-------------] 2/6
Transfer   [###################-----] 39          Active  [--------------------] 0 idle=2
Withdrawal [##----------------------] 5           Queue   waits=0 avg=0.0ms/op=0.00 lock=0

Validation                                        Errors
Final audit healthy                               No operation errors
Check        Expected     Actual Status
---------- ---------- ---------- ------
Accounts            6          6 ok
Txns               55         55 ok
Entries           110        110 ok
External           16         16 ok
Reversals           0          0 clear
Issues              0          0 clear

Final
  grade:     A/95 (correct, fast, and stable)
  signals:   rps=720.4 p95=6.10ms p99=10.30ms pool_wait/op=0.00ms lock_waiters_max=0
  mix:       deposit=6 transfer=39 withdrawal=5
```

Use `-format=json` for newline-delimited JSON events:

- `progress`: operation counts, success/failure counts, error counts, latency,
  throughput, retry attempts, database pool stats, and the latest in-run
  validation result
- `validation`: same shape as progress, emitted after optional in-run validation
- `final`: final stats, database pool stats, and the full validator result

Useful fields:

- `stats.total`
- `stats.success`
- `stats.failure`
- `stats.ops_per_second`
- `stats.max_lock_waiters`
- `stats.latency_ms.p95`
- `stats.latency_ms.p99`
- `stats.by_operation`
- `stats.by_error_code`
- `stats.by_error_category`
- `db_stats.wait_count`
- `db_stats.wait_duration_ms`
- `summary.performance.grade`
- `summary.performance.pool_wait_per_op_ms`
- `summary.final_validation.healthy`
- `summary.final_validation.summary.issue_count`

## Example Commands

Small smoke run:

```bash
go run ./cmd/stress \
  -accounts=5 \
  -workers=2 \
  -operations=50 \
  -think-min=50ms \
  -think-max=250ms
```

Larger local run with periodic validation:

```bash
go run ./cmd/stress \
  -accounts=50 \
  -workers=20 \
  -operations=5000 \
  -think-min=25ms \
  -think-max=500ms \
  -validation-interval=5s
```

The runner defaults `-max-open-conns` to `min(workers + 4, 50)`. That keeps a
large worker count from exhausting a local Postgres instance. Increase it only
when the database is configured to accept more connections.

The default `-operation-timeout` is `30s` so large local runs can wait behind the
database pool and row locks without producing misleading timeout failures.

Settlement bucket run:

```bash
go run ./cmd/stress \
  -accounts=500 \
  -workers=150 \
  -operations=50000 \
  -settlement-buckets=16
```

Hot-account transfer run:

```bash
go run ./cmd/stress \
  -accounts=500 \
  -workers=150 \
  -operations=50000 \
  -hot-accounts=5 \
  -hot-transfer-percent=20
```

Use settlement buckets to test whether a single internal clearing row is driving
tail latency. Use hot-account mode to test skewed wallet-to-wallet traffic.
This command is intentionally rough: 20% of transfers hit only 5 wallet rows, so
large P95/P99 values are expected if workers outpace the database pool or row
locks.

Pool comparison runs:

```bash
for conns in 25 50 75 100; do
  go run ./cmd/stress \
    -accounts=500 \
    -workers=150 \
    -operations=50000 \
    -max-open-conns="$conns"
done
```

Compare `rps`, `p95`, `p99`, `pool_wait/op`, and `lock_waiters_max`. If more
connections reduce pool waits but not tail latency, the bottleneck is probably
row contention or write duration rather than connection admission.

JSON output for dashboards or scripts:

```bash
go run ./cmd/stress \
  -accounts=25 \
  -workers=10 \
  -operations=1000 \
  -format=json
```

The runner exits non-zero if final validation fails or if unexpected non-business
errors occurred. Business errors such as insufficient funds are counted as
operation failures but do not fail the run by themselves.

## Next Runners

Add these only after the direct Go runner is useful:

- HTTP runner: drives `/transfers`, `/external-transfers`, and `/validation` to
  include service serialization, JSON parsing, and HTTP timeout behavior.
- External dashboard: reads progress JSON or polls `/validation` and renders
  longer-running history outside the terminal.

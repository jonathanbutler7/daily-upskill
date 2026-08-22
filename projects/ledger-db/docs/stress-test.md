# Stress Test Plan

The first stress-test runner is a local Go CLI:

```bash
go run ./cmd/stress
```

Use `go run ./cmd/stress --help` for quick commands. Use
`go run ./cmd/stress --help-full` for the grouped flag reference. This doc owns
the longer examples, dashboard fields, and interpretation notes.

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
- optional concurrent duplicate requests with the same idempotency key
- optional mismatched-key requests that should return idempotency conflicts
- retry behavior for retryable database errors
- randomized operation mix, account choice, amount, and worker pacing
- final committed-state validation through `internal/ledgervalidator`

This runner is for write-path and database contention. It does not measure HTTP
server overhead, and it does not create alternate settlement accounts. External
transfers use the single `Cash Settlement` account seeded by the migrations.

## Visibility

The runner prints a live terminal dashboard by default:

```text
LedgerDB Status  Health RUNNING  Run d17edc05  18:10:42  accounts=5 workers=2 ops=1000
Local Postgres . USD ledger . randomized deposits, withdrawals, and wallet transfers

Workload                                          Goroutines
Ops      [###########-----------------]  40.0%    Workers  2 worker goroutines
OK       [###########-----------------]  20       Runtime  8 goroutines now
Failed   [----------------------------]  0        Other    6 non-worker goroutines
RPS      [##################] 12.5/sec            Report   1 dashboard/validation
Elapsed  4s retries=0                             DB cap   max_open_conns=6

Operations                                        DB Pool
Deposit    [###---------------------] 6           Held    [#######-------------] 2/6
Transfer   [###################-----] 39          Active  [--------------------] 0 idle=2
Withdrawal [##----------------------] 5           Waited  0 pool waits avg=0.0ms/op=0.00
Spread     accounts/worker=2.5 hot=off           Failed  pool_timeouts=0 server_rejects=0 lock=0

Idempotency                                      Concurrency
Requests  1008 total for 1000 logical ops        Workers  2 db_conns=6
Duplicates 6 extra requests in 3 batches         Replays  6 same-transaction returns
Conflicts  2 expected mismatched-key rejects      Unexpected [------------------] 0

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
  signals:   rps=720.4 worker_goroutines=2 runtime_goroutines=8 pool_wait/op=0.00ms pool_timeouts=0 server_rejects=0 lock_waiters_max=0
  mix:       deposit=6 transfer=39 withdrawal=5
  idem:      requests=58 duplicate_extra=6 replays=6 conflicts=2 unexpected=0
```

The `#` bars use different scales by section:

- `Workload` bars show progress toward the requested operation count.
- `Goroutines` shows configured worker goroutines, current runtime goroutines, and non-worker goroutines.
- `DB Pool` bars show usage against `-max-open-conns`.

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

Larger local run:

```bash
go run ./cmd/stress \
  -accounts=50 \
  -workers=20 \
  -operations=5000 \
  -think-min=25ms \
  -think-max=500ms
```

The runner defaults `-max-open-conns` to `min(workers + 4, 50)`. That keeps a
large worker count from exhausting a local Postgres instance. Increase it only
when the database is configured to accept more connections.

The CLI uses fixed defaults for seed balance, amount size, operation timeout,
reset behavior, and output format.

Hot-account transfer run:

```bash
go run ./cmd/stress \
  -accounts=500 \
  -workers=150 \
  -operations=50000 \
  -hot-accounts=5 \
  -hot-transfer-percent=20
```

Use hot-account mode to test skewed wallet-to-wallet traffic. This command is
intentionally rough: 20% of transfers hit only 5 wallet rows, so visible pool
waits and lock waiters are expected if workers outpace the database. If the
single settlement account becomes the bottleneck, the stress runner should
surface that pressure rather than working around it.

Concurrent duplicate request run:

```bash
go run ./cmd/stress \
  -accounts=25 \
  -workers=20 \
  -operations=1000 \
  -duplicate-percent=10 \
  -duplicate-fanout=3 \
  -conflict-percent=2
```

`-duplicate-percent` chooses how often a logical operation fans out same-key
requests at the same time. `-duplicate-fanout=3` means one logical operation
sends three concurrent requests with the same operation fields and idempotency
key. The committed transaction count should still rise by one for that logical
operation.

`-conflict-percent` follows successful operations with a same-key request that
changes the amount. Those requests should be rejected as idempotency conflicts.
In a healthy run, `idempotent_replays` and `idempotency_conflicts` rise while
`idempotency_unexpected` stays at zero.

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

Compare `rps`, `pool_wait/op`, `pool_timeouts`, `server_rejects`, and
`lock_waiters_max`. If more connections reduce pool waits but lock waiters stay
high, the bottleneck is probably row contention or write duration rather than
connection admission.

The runner exits non-zero if final validation fails or if unexpected non-business
errors occurred. Business errors such as insufficient funds are counted as
operation failures but do not fail the run by themselves.

## Next Runners

Add these only after the direct Go runner is useful:

- HTTP runner: drives `/transfers`, `/external-transfers`, and `/validation` to
  include service serialization, JSON parsing, and HTTP timeout behavior.
- External dashboard: polls `/validation` and renders longer-running history
  outside the terminal.

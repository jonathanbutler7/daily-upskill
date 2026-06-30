# Reliability Patterns

Making payment systems work correctly even when things fail.

## Core Concepts

### Idempotency

An operation is idempotent if running it multiple times produces the same result as running it once.

Payment APIs must be idempotent. If a network timeout happens after you send a charge request, you don't know if it succeeded. You'll retry. Without idempotency, you might charge the customer twice.

**Implementation**: Client sends an idempotency key with each request. Server checks if it's seen that key before. If yes, return the original result. If no, process the request and store the result.

```
POST /charges
Idempotency-Key: abc-123
{amount: 100}

# First call: process charge, store result keyed by "abc-123"
# Second call: return stored result, don't charge again
```

### Safe Retries

Not all operations are safe to retry:

- **Safe**: GET requests, idempotent writes
- **Unsafe**: Non-idempotent writes, operations with side effects

Before retrying, ask: "If this already succeeded, will retrying cause harm?"

### Webhook Handling

Processors send webhooks to notify you of events. Problems to handle:

**Deduplication**: Same webhook sent twice. Store webhook IDs, skip duplicates.

**Ordering**: `payment.captured` arrives before `payment.authorized`. Either buffer and reorder, or make your handlers tolerate out-of-order events.

**Replay**: Webhook failed, processor retries hours later. Your system should handle stale events gracefully.

**Acknowledgment**: Return 200 quickly. Do heavy processing async. If you timeout, the processor will retry.

### Outbox Pattern

Problem: You need to update your database AND send a message to another service. If the DB write succeeds but the message fails, you have inconsistent state.

Solution: Write the message to an "outbox" table in the same transaction as your data. A separate process reads the outbox and sends messages. If sending fails, retry from the outbox.

```sql
BEGIN;
  INSERT INTO payments (id, amount) VALUES (1, 100);
  INSERT INTO outbox (event_type, payload) VALUES ('payment.created', '{"id": 1}');
COMMIT;
```

### State Machines

Model payment lifecycle as explicit states and transitions. Benefits:

- Invalid transitions are impossible (can't refund a declined payment)
- Current state is always clear
- Audit log of transitions
- Easier to reason about edge cases

```go
type PaymentState string
const (
    Created   PaymentState = "created"
    Authorized PaymentState = "authorized"
    Captured  PaymentState = "captured"
    Refunded  PaymentState = "refunded"
)

func (p *Payment) CanTransitionTo(next PaymentState) bool {
    allowed := map[PaymentState][]PaymentState{
        Created:    {Authorized},
        Authorized: {Captured},
        Captured:   {Refunded},
    }
    for _, s := range allowed[p.State] {
        if s == next { return true }
    }
    return false
}
```

### Sagas and Compensating Actions

When a multi-step process fails partway through, you need to undo the completed steps.

Example: Reserve inventory → Charge card → Ship order

If shipping fails, you need to:
1. Refund the charge (compensating action)
2. Release the inventory (compensating action)

Each step needs a corresponding undo operation.

## What Can Go Wrong

- **Missing idempotency** - Customer charged twice
- **Unsafe retry** - Side effect happens multiple times
- **Lost webhook** - State never updates, payment stuck
- **Outbox not drained** - Messages pile up, downstream never notified
- **State machine bypass** - Code allows invalid transition, data corrupted
- **Incomplete saga** - Failure mid-process, no compensation, inconsistent state

## How This Connects to My Projects

**payer-sync**: Uses idempotency keys for payment processing. State machine for lifecycle. Audit log for every transition. Retry logic for transient failures.

**ledger-db**: Idempotency via unique constraint on transaction key. Atomic writes via Postgres function. Immutable entries (compensation via reversal, not edit).

## Questions to Ask About Real Systems

- How do we handle duplicate requests?
- What's our retry policy for failed operations?
- How do we dedupe webhooks?
- Do we use an outbox pattern for cross-service communication?
- How do we handle partial failures in multi-step processes?

## Keywords

Idempotency, idempotency key, safe retry, webhook deduplication, outbox pattern, inbox pattern, state machine, saga, compensating action, at-least-once delivery, exactly-once semantics

# Payment State

Understanding the difference between what the customer sees, what your system thinks, and what actually happened.

## Core Concepts

### Multiple Sources of Truth

A payment has state in multiple places:

- **Customer-visible state** - What the UI shows ("Payment successful!")
- **Application state** - What your database says
- **Processor state** - What Stripe/Adyen thinks
- **Bank/network state** - What actually happened with the money

These don't always agree. A lot of payment bugs come from treating one as proof of the others.

### Payment Lifecycle States

Typical states a payment moves through:

```
created → pending → authorized → captured → settled → funded
                 ↘ declined
                 ↘ expired
        captured → refunded
        captured → disputed → chargeback_won / chargeback_lost
```

### Terminal vs Non-Terminal States

- **Terminal**: Payment is done. No more transitions. (succeeded, failed, refunded, chargeback_lost)
- **Non-terminal**: Payment is still in progress. (pending, authorized, disputed)

Knowing which states are terminal matters for retries, reconciliation, and reporting.

### Webhooks and Events

Processors send webhooks when state changes. Problems:

- **Delayed** - Webhook arrives minutes or hours after the event
- **Out of order** - `captured` webhook arrives before `authorized`
- **Duplicate** - Same event sent twice
- **Missing** - Webhook never arrives (network failure, misconfiguration)

Your system needs to handle all of these.

## What Can Go Wrong

- **Stale state** - Your DB says "pending" but the processor says "captured" because you missed a webhook
- **Race conditions** - User clicks "pay" twice, two authorizations created
- **Zombie payments** - Payment stuck in non-terminal state forever because webhook never arrived
- **State machine violations** - Code allows transitions that shouldn't be possible (refunding a declined payment)

## How This Connects to My Projects

**payer-sync**: The lifecycle statuses (RECEIVED_RAW → PARSED → MATCHED → PROCESSING_PAYMENT → POSTED) are a state machine. Each transition is audited. Invalid transitions are blocked.

**ledger-db**: Ledger entries are immutable. You don't change state by editing; you add new entries that represent the transition (reversal entries, adjustment entries).

## Questions to Ask About Real Systems

- What states can a payment be in?
- Which states are terminal?
- How do we handle out-of-order webhooks?
- How do we detect stuck payments?
- What's our webhook retry policy?
- How do we dedupe duplicate webhooks?

## Keywords

Payment state machine, terminal state, webhook, event-driven, idempotency, out-of-order events, duplicate events, state transition, lifecycle

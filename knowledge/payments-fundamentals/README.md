# Payments Fundamentals

Core concepts for understanding how money moves, what can go wrong, and how to prove correctness.

## Why This Matters

Payments is the domain I'm building expertise in for Staff Engineer. The goal is to become the engineer people pull into design conversations when money movement, correctness, or risk is involved.

The useful version of payments expertise is practical:
- You know how money moves
- You know where system state can lie
- You know what can be retried safely
- You know how to prove what happened later
- You know who gets hurt when the system is wrong

## Topics

| Topic | Description | Status |
|-------|-------------|--------|
| [Money Movement](money-movement.md) | Authorization, capture, settlement, funding | ✅ Complete |
| [Payment State](payment-state.md) | Lifecycle, state machines, webhooks | ✅ Complete |
| [Ledger & Accounting](ledger-accounting.md) | Double-entry, debits/credits, immutability | ✅ Complete |
| [Reconciliation](reconciliation.md) | Matching records, exceptions, corrections | ✅ Complete |
| [Reliability Patterns](reliability-patterns.md) | Idempotency, retries, sagas, outbox | ✅ Complete |
| [Risk & Compliance](risk-compliance.md) | Fraud, KYC, PCI, chargebacks | ✅ Complete |

## How to Use This

Each topic has:
- Core concepts (what you need to know)
- How it connects to projects I've built
- Questions to ask about real systems
- Keywords for further research

These are reference docs, not tutorials. They assume I've done the work to understand the concept.

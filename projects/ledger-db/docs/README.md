# Docs

This directory contains notes and decisions for the ledger project.

## Concepts

1. Posted entries balance
2. Atomic posting
3. Idempotency
4. No negative balance
5. Isolation
6. Reversal
7. Store Balance vs. Derived Balance

## Scenarios

Examples of 
- balances
- transactions
- entries
- invariants

## Schema

The beginnings of an ERD with fields and relationships sketched out.

## System Boundaries

See [system-boundaries.md](system-boundaries.md) for the current decision on what belongs in Go and what belongs in Postgres.

Short version:

- Postgres owns durable ledger invariants.
- Go owns commands, workflows, retries, and caller-facing errors.
- Some rules exist in both places, but for different reasons.

### Out of scope

- users
- entity states

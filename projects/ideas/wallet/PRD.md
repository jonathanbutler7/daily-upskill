# Wallet PRD

## Summary

Build `wallet`, a small stored-value wallet product that lets users hold a balance, add money, send money to another user, withdraw money, and view a clear transaction history.

The project should integrate with `ledger-db`. `wallet` owns the product API, user flows, limits, and transaction history. `ledger-db` owns double-entry accounting, balances, idempotency, reversals, and external-transfer records.

This is a strong follow-on project because it turns `ledger-db` from a backend accounting exercise into a product-facing money movement system.

## Problem

`ledger-db` can model balances and transfers, but it does not yet have a realistic product sitting in front of it. Real wallets need more than ledger entries. They need account creation, funding flows, withdrawals, user-to-user transfers, transaction status, limits, idempotent API behavior, and a clean history users can understand.

Without a wallet layer, it is hard to practice the boundary between:

- a user action
- a product transaction
- an external money movement attempt
- an internal ledger posting
- a failed or reversed transfer
- a user-facing transaction history item

`wallet` should provide that product layer.

## Goals

- Let a user create a wallet account.
- Let a user add funds from an external source.
- Let a user send funds to another wallet user.
- Let a user withdraw funds to an external destination.
- Show available and posted balance.
- Show transaction history with simple user-facing statuses.
- Use idempotency keys for all money-moving requests.
- Use `ledger-db` as the accounting source of truth.
- Keep product state separate from ledger accounting state.
- Leave a clear path for later integration with `rail-sim`.

## Non-Goals

- Real bank account linking.
- Real card processing.
- Real KYC, AML, sanctions screening, or fraud operations.
- Custody, licensing, or compliance implementation.
- Multi-currency wallets in the MVP.
- Interest, rewards, subscriptions, or merchant payments.
- Mobile app build.
- Crypto wallet behavior.

## Users

### Primary User

A consumer with a simple stored-value wallet.

They want to add money, send money, withdraw money, and know what happened.

### Secondary User

The developer building payments systems knowledge.

They need a product layer that exercises `ledger-db` through realistic flows: funding, peer transfer, withdrawal, idempotency, insufficient funds, failed external transfers, and reversals.

### Reviewer

An interviewer or technical reviewer.

They should be able to see that the project separates product workflows from ledger accounting and external rail state.

## Core Concepts

### Wallet Account

A user-owned stored-value account.

Example fields:

- `wallet_id`
- `user_id`
- `ledger_account_id`
- `status`
- `created_at`

### Wallet Transaction

The product-facing record shown to a user.

Example fields:

- `wallet_transaction_id`
- `wallet_id`
- `type`
- `amount`
- `currency_code`
- `status`
- `ledger_transaction_id`
- `external_transfer_id`
- `created_at`

### Funding

Money added from an external source into a wallet.

In the MVP, the external source can be fake. The important behavior is that the wallet records the request and posts to `ledger-db` only when the funding event should affect balance.

### Peer Transfer

A transfer from one wallet user to another.

This should be posted as a balanced internal ledger transaction:

```text
sender wallet   -amount
receiver wallet +amount
```

### Withdrawal

Money moved out of a wallet to an external destination.

This should create product state, external-transfer state, and balanced ledger entries once the withdrawal should affect the wallet balance.

## MVP Product Surface

`wallet` should expose an HTTP API and store product state in Postgres.

```text
POST /users
POST /wallets
GET /wallets/{wallet_id}
GET /wallets/{wallet_id}/balance
GET /wallets/{wallet_id}/transactions

POST /wallets/{wallet_id}/fund
POST /wallets/{wallet_id}/withdraw
POST /wallets/{wallet_id}/transfers
GET /wallet-transactions/{wallet_transaction_id}
```

All money-moving POST requests must require an idempotency key.

## Transaction States

```text
requested
pending
posted
failed
reversed
```

Product transaction status should be understandable to a user. It does not need to expose every ledger or rail detail directly.

## Ledger DB Integration

Yes, this project should integrate with `ledger-db`.

The clean boundary:

- `wallet` owns users, wallet accounts, product transactions, limits, and user-facing history.
- `ledger-db` owns ledger accounts, ledger transactions, ledger entries, balances, idempotency, reversals, and external-transfer records.
- A wallet account maps to one `ledger-db` ledger account.
- A wallet funding or withdrawal maps to a `ledger-db` external transfer.
- A peer transfer maps to a `ledger-db` internal transfer.

Example funding request:

```json
{
  "wallet_id": "wal_123",
  "amount": 5000,
  "currency_code": "USD",
  "external_reference": "fund_001",
  "idempotency_key": "idem_001"
}
```

Expected `ledger-db` behavior:

- Funding success: post `Cash Settlement -> user wallet`.
- Withdrawal success: post `user wallet -> Cash Settlement`.
- Peer transfer: post `sender wallet -> receiver wallet`.
- Duplicate request: return the original result without double-posting.
- Failed external transfer: keep product status failed and avoid posting settled money.
- Reversal: create a correcting ledger transaction instead of editing history.

## MVP Scenarios

1. User creates a wallet and sees a zero balance.
2. User funds a wallet and sees balance increase.
3. User retries the same funding request and does not get double credit.
4. User sends money to another wallet user.
5. User tries to send more than available balance and gets a clear failure.
6. User withdraws funds and sees balance decrease.
7. A funding or withdrawal fails before posting.
8. A posted transaction is reversed through `ledger-db`.

## Success Criteria

- A developer can run the wallet service locally.
- A user can create a wallet, fund it, send money, withdraw money, and view history through the API.
- Every posted wallet movement has a matching balanced transaction in `ledger-db`.
- Duplicate money-moving requests are idempotent.
- Insufficient funds are rejected before a wallet balance goes negative.
- Product transaction history can be rebuilt or verified from `ledger-db` references.
- At least one failure path and one reversal path are covered by tests.

## Open Questions

- Should `wallet` live as a separate service or as a client package inside `ledger-db` first?
- Should funding post immediately in the MVP, or should it wait for a simulated settled event?
- Should available balance differ from posted balance in the first version?
- Should wallet transaction history be stored directly, projected from ledger events, or both?
- Should the next integration target be `ledger-db` directly or `rail-sim` plus `ledger-db` together?

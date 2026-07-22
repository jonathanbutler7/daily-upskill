# Rail Sim PRD

## Summary

Build `rail-sim`, a sandbox payment rail simulator that mimics the role of an ACH operator such as FedACH or EPN. The app should accept payment instructions from a fake originating bank, validate and process them through simulated clearing windows, route them to fake receiving banks, and emit settlement or return events.

`rail-sim` is a small system for making delayed settlement, returns, reconciliation, and external rail state easy to understand and test.

`rail-sim` should be able to integrate with `ledger-db` as an external rail. `rail-sim` owns the simulated rail lifecycle. `ledger-db` owns internal accounting.

## Problem

`ledger-db` can record external transfers, but it does not yet have a realistic source of rail events. Real payment systems deal with files, windows, effective dates, settlement dates, returns, reversals, duplicate references, and reconciliation reports.

Without a rail simulator, it is hard to practice the difference between:

- a transfer request
- an accepted rail instruction
- a settled payment
- a returned payment
- an internal ledger posting
- a reconciliation exception

`rail-sim` should provide that missing outside system.

## Goals

- Simulate ACH-style clearing and settlement without connecting to any real bank or network.
- Accept credit and debit entries from a fake ODFI or fintech client.
- Validate basic entry fields before accepting them.
- Run simulated processing windows.
- Route entries to fake RDFIs by routing number.
- Track entry state from submitted to accepted, settled, returned, or rejected.
- Emit events that `ledger-db` can consume using stable external references.
- Produce settlement and reconciliation reports.

## Non-Goals

- Real ACH network access.
- Real Nacha file certification.
- Real bank account verification.
- Real money movement.
- Full Nacha rule coverage.
- Consumer-facing UI.
- Fraud/risk scoring beyond simple scripted scenarios.
- RTP, card network, wire, or check support in the MVP.

## Users

### Primary User

The developer building payments knowledge through `ledger-db`.

They need a realistic external system that can create delayed success, delayed failure, duplicate events, and settlement reports.

### Secondary User

An interviewer or reviewer looking at the project.

They should be able to see that the system models the boundary between internal ledger accounting and external payment rail state.

## Core Concepts

### ODFI

The originating bank or sponsor bank that submits ACH entries into the simulator.

### RDFI

The receiving bank that receives routed entries and may accept, settle, or return them.

### Entry

A single credit or debit instruction.

Example:

```json
{
  "external_reference": "ach_001",
  "direction": "credit",
  "amount": 200000,
  "currency_code": "USD",
  "odfi_routing_number": "111000025",
  "rdfi_routing_number": "222000033",
  "receiver_account_number": "123456789",
  "effective_date": "2026-07-21"
}
```

### Processing Window

A simulated batch run that moves eligible entries forward. Entries submitted after a cutoff should wait for the next window. Future-dated entries should wait until their effective date.

### Settlement Report

A report showing net settlement positions for each participating bank on a given settlement date.

## MVP Product Surface

`rail-sim` should expose an HTTP API and store state in Postgres or SQLite.

### API

```text
POST /entries
GET /entries/{id}
GET /entries?external_reference=ach_001
POST /operator/windows/run
GET /rdfis/{routing_number}/incoming
POST /entries/{id}/return
GET /settlement/{date}
GET /events
```

### Entry States

```text
submitted
accepted
rejected
pending_settlement
settled
returned
```

### Event Types

```text
entry.accepted
entry.rejected
entry.pending_settlement
entry.settled
entry.returned
settlement.report_created
```

Events must be idempotent and include:

- event ID
- event type
- external reference
- amount
- currency
- direction
- current rail status
- effective date
- settlement date, when known
- created timestamp

## Ledger DB Integration

`ledger-db` should treat `rail-sim` as a payment rail, not as part of the ledger.

The basic contract:

```json
{
  "event_id": "evt_001",
  "event_type": "entry.settled",
  "external_reference": "ach_001",
  "rail": "ach",
  "direction": "deposit",
  "status": "settled",
  "amount": 200000,
  "currency_code": "USD",
  "effective_date": "2026-07-21",
  "settled_at": "2026-07-21T14:00:00Z"
}
```

Expected `ledger-db` behavior:

- Accepted ACH entry: create or update external transfer state, but do not assume final settlement.
- Settled ACH credit: post balanced ledger entries against `Cash Settlement`.
- Returned ACH entry: mark the external transfer failed or create a correcting ledger transaction if money was already posted.
- Duplicate event: process idempotently.
- Amount/reference mismatch: create a reconciliation exception.

## MVP Scenarios

1. ACH credit settles successfully.
2. ACH debit settles successfully.
3. Entry is rejected before processing because required fields are invalid.
4. Entry is accepted but later returned.
5. Entry misses the current processing window and settles later.
6. Entry has a duplicate external reference.
7. Settlement report does not match `ledger-db` external transfer state.

## Success Criteria

- A developer can submit a fake ACH credit and see it move through accepted, pending settlement, and settled.
- A developer can submit a fake ACH debit and force a return.
- The app emits stable events that can be replayed safely.
- Settlement reports show net positions by bank and date.
- `ledger-db` can consume at least one settled event and create a matching `external_transfers` row.
- A reconciliation check can detect at least one mismatch between rail state and ledger state.

## Open Questions

- Should the MVP use SQLite for simplicity or Postgres to match `ledger-db`?
- Should `rail-sim` push events to `ledger-db`, or should `ledger-db` poll `/events`?
- Should the first integration post only settled events, or should it model pending external transfers first?
- How much Nacha file format should be simulated before it becomes distracting?

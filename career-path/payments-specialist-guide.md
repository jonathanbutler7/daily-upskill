# Payments Specialist Guide

**Goal**: Become strong enough in payments that people pull you into design conversations early.

Payments is a good Staff-level domain because the work crosses backend systems, data correctness, risk, compliance, finance, product, and operations.

The useful version of payments expertise is practical:

- You know how money moves.
- You know where system state can lie.
- You know what can be retried safely.
- You know how to prove what happened later.
- You know who gets hurt when the system is wrong.

---

## Payments Fundamentals

### Money Movement

Learn how funds move between customers, merchants, processors, acquiring banks, issuing banks, card networks, ACH operators, and bank accounts.

Start with:

- Authorization
- Capture
- Clearing
- Settlement
- Funding
- Payout
- Refund
- Reversal
- Dispute
- Chargeback

### Payment State

Learn the difference between:

- Customer-visible state
- Internal application state
- Processor state
- Bank or network state
- Ledger state

These do not always update at the same time. A lot of payment bugs come from treating one of these as if it proves the others.

### Ledger and Accounting

The term is **double-entry bookkeeping** or **double-entry accounting**.

Every financial event creates entries that balance. Debits and credits do not simply mean "money in" and "money out." Their meaning depends on the account type.

Learn:

- Accounts
- Transactions
- Ledger entries
- Debits and credits
- Pending vs posted entries
- Derived balances
- Reversals
- Holds and reserves
- Fees
- Payouts
- Immutability

The big rule: do not edit financial history. Add new entries that correct it.

### Reconciliation

Reconciliation is how you find out whether your system agrees with outside reality.

Learn:

- What external reports exist
- What fields are matched
- What timing differences are normal
- Which mismatches mean money is at risk
- Who reviews exceptions
- How corrections are made

### Reliability

Payments systems need boring, careful reliability work:

- Idempotency keys
- Safe retries
- Webhook dedupe
- Webhook replay
- State machines
- Outbox/inbox patterns
- Audit logs
- Manual recovery paths
- Exception queues

### Risk and Compliance

You do not need to become a compliance expert, but you should know enough to ask better questions.

Learn the basics of:

- PCI
- Tokenization
- KYC/KYB
- AML
- Sanctions
- Fraud review
- Disputes
- Data retention
- Sensitive-data logging

### Business Mechanics

Payments choices show up in business metrics.

Learn:

- Authorization rate
- Decline codes
- Processor fees
- Interchange
- Settlement timing
- Cash flow
- Refund cost
- Dispute cost
- Support load

---

## How to Learn From Work

Pick one real payment flow and trace it all the way through.

Answer:

- What starts the payment?
- Which services touch it?
- Which external systems receive requests?
- What IDs exist at each layer?
- What states can it enter?
- Which states are terminal?
- What can be retried?
- What needs manual review?
- What does the customer see?
- What does Finance or Operations see?

Then write one artifact:

- Sequence diagram
- State machine
- Failure-mode list
- Ledger example
- Reconciliation map
- Runbook

Do this once a month. Keep it small enough to actually finish.

---

## Good Projects

### 1. Double-Entry Ledger

This is the best next project.

Build a small ledger inspired by Modern Treasury's ledger concepts. The goal is to make double-entry bookkeeping concrete.

Build:

- Accounts
- Transactions
- Ledger entries
- Debits and credits
- Balance queries
- Pending and posted entries
- Reversals
- Idempotency keys
- External payment IDs
- Reconciliation against an imported processor or bank report

Test it with:

- Payment
- Processor fee
- Refund
- Chargeback
- Chargeback reversal
- Payout
- Duplicate webhook
- Processor success with internal write failure

### 2. Card Payment Lifecycle Simulator

Model auth, capture, partial capture, void, refund, dispute, settlement, and payout.

This helps with payment state and timing differences between your system and the processor.

### 3. ACH Return Simulator

Model ACH debit origination, settlement windows, returns, reversals, and customer-visible state.

This teaches delayed failure. ACH "success" can be provisional.

### 4. Webhook Inbox and Replay Tool

Build webhook ingestion with:

- Signature validation
- Duplicate detection
- Out-of-order event handling
- Replay
- Idempotent state changes

### 5. Settlement Reconciliation Tool

Import fake processor settlement reports and compare them against internal records.

Report:

- Matched records
- Missing records
- Duplicates
- Delayed records
- Amount mismatches
- Fee mismatches

### 6. Payment Risk Review Queue

Use the payment-likelihood ML project as input to a review workflow.

The useful parts are:

- Thresholds
- Review reasons
- Audit notes
- False positives
- False negatives
- Monitoring
- Safe use of the score

---

## Questions to Ask

Use these when reading code or reviewing a design:

- What is the source of truth?
- What external system can disagree with us?
- What happens if this times out?
- Can this run twice safely?
- What ID links our record to the processor, bank, network, or ledger?
- Is this customer-visible state or internal state?
- What states are terminal?
- Who owns the financial loss if this fails?
- How do we know something is stuck?
- How do we reconcile this later?
- How would Operations fix it?
- What audit trail proves what happened?
- What sensitive data are we touching?
- What metric tells us this is healthy?

---

## Signs of Progress

You are getting stronger when:

- You can explain a payment from customer action to settlement.
- You can tell the difference between internal success and external financial confirmation.
- You naturally ask about idempotency, reconciliation, auditability, and recovery.
- You can debug with both logs and payment concepts.
- You can explain ledger entries for payment, fee, refund, dispute, and payout.
- You can spot payment-state ambiguity in a design.
- You have shipped or documented something that reduces money risk, manual work, or customer confusion.

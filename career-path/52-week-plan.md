# 52-Week Payments Specialist Plan

**Goal**: Build Staff-level technical authority in payments  
**Timeline**: May 4, 2026 - May 3, 2027  
**Current date for this plan**: June 17, 2026  
**Current track**: Payments fundamentals, ledger correctness, reconciliation, risk, and reliable money movement

This plan changed because the pay syncing project is done. PayerSync still counts as proof that you can take a payments-adjacent system from idea to documented design, but it should not consume the rest of the year.

The next phase should be more direct: learn payments fundamentals, build projects that force the concepts to become concrete, and connect that work back to real payment systems at work.

---

## Completed Work So Far

### Pay Syncing Project

**Status**: Done around May 27, 2026.

This project covered ingestion, reconciliation, payment processing boundaries, writeback, notifications, auditability, and failure handling. Keep it as evidence for:

- System boundary thinking
- Reconciliation and source-of-truth questions
- Idempotency and retry design
- Failure-mode analysis
- Technical writing and diagrams

### Payment Propensity ML Project

**Status**: Done June 9-10, 2026.

This project predicted the likelihood that a person will pay. Document it as a short case study, not as the main career direction. The value is that it connects payments behavior, data modeling, and product/business decisions.

Document:

- What question the model answered
- What inputs/features were used
- What target label meant
- What the model predicted
- How the prediction could be used safely
- What risks exist: bias, overconfidence, stale data, false positives, false negatives
- What you would monitor if this ran in production

---

## Payments Fundamentals

Payments fundamentals are the concepts you need to reason about money movement, system state, risk, and correctness.

### 1. Money Movement

- Authorization, capture, clearing, settlement, funding, and payout
- Card payments, ACH, RTP, wires, wallets, and checks
- Issuers, acquirers, processors, gateways, networks, payment facilitators, and merchants
- Timing differences between customer-visible success and actual funds movement

### 2. Payment Lifecycle State

- Payment intent, authorization, capture, cancellation, refund, reversal, dispute, chargeback, and payout
- Pending, succeeded, failed, settled, returned, reversed, disputed, and terminal states
- Internal state versus processor state versus bank/network state
- Webhooks, polling, delayed events, duplicate events, and out-of-order events

### 3. Ledger and Accounting

The term you want is **double-entry bookkeeping** or **double-entry accounting**.

The key idea: every financial event creates at least two ledger entries, and the entries must balance. Debits and credits are accounting terms. They do not simply mean "money in" and "money out." Their meaning depends on the account type.

You should learn:

- Accounts, balances, entries, transactions, debits, and credits
- Assets, liabilities, revenue, expenses, and contra accounts
- Pending versus posted ledger entries
- Holds, reserves, fees, refunds, chargebacks, payouts, and adjustments
- Immutability: correcting bad entries with reversal entries instead of editing history
- Invariants: every transaction balances, account balances derive from entries, and history is auditable

### 4. Reconciliation

- Matching internal records to processor reports, bank reports, and partner records
- Handling expected timing differences versus true mismatches
- Exception queues, manual review, correction workflows, and audit notes
- Settlement reports, payout reports, fees, disputes, refunds, and returns

### 5. Reliability Patterns

- Idempotency keys
- Safe retries
- Webhook dedupe and replay
- Outbox/inbox patterns
- State machines
- Sagas and compensating actions
- Partial failure handling
- Audit logs and operational recovery tools

### 6. Risk, Fraud, and Compliance

- Fraud scoring, velocity checks, identity checks, and manual review
- KYC, KYB, AML, sanctions, PCI, privacy, and data retention
- Disputes, chargebacks, returns, losses, reserves, and risk ownership
- What data should never be logged or exposed

### 7. Business Mechanics

- Authorization rate
- Decline codes
- Interchange and processor fees
- Settlement timing
- Cash flow
- Refund and dispute cost
- Customer trust and support load

---

## Project Ideas

### Recommended Main Project: Double-Entry Ledger

Build a small ledger inspired by Modern Treasury's ledger concepts. This is the best next project because ledger correctness is foundational for payments expertise.

Build:

- Accounts
- Transactions
- Ledger entries
- Debits and credits
- Balance queries
- Pending and posted entries
- Reversals
- Idempotency keys
- Metadata linking ledger entries to external payment IDs
- A reconciliation report that compares expected balances to imported processor or bank records

Good test cases:

- Customer payment
- Processor fee
- Refund
- Failed refund
- Chargeback
- Chargeback reversal
- Payout to merchant
- Duplicate webhook
- Processor success but internal write failure

### Other Strong Projects

- **Card lifecycle simulator**: model auth, capture, partial capture, void, refund, dispute, and settlement.
- **ACH return simulator**: model ACH debit creation, settlement window, returns, reversals, and customer-visible state.
- **Webhook inbox and replay tool**: ingest duplicate/out-of-order processor webhooks and safely drive a payment state machine.
- **Settlement reconciliation tool**: import fake processor reports and find missing, duplicated, delayed, or fee-mismatched records.
- **Payment risk review queue**: extend the ML project into a safer operational workflow with thresholds, review reasons, and monitoring.
- **Decline analytics dashboard**: group failed payments by decline code, issuer, amount range, retry count, and payment method.

---

## How to Use This Plan

Each week has three parts:

- **Artifact**: a repo doc, diagram, test, small implementation, or case study.
- **Work connection**: one way to connect the topic to a real payment system at work.
- **Done when**: a concrete stopping point.

Use the weekly plan as a track, not a prison. If a real work payments project appears, it can replace the toy project for that week as long as it teaches the same concept.

---

## Q1: Reset the Track and Build the Payments Map (Weeks 1-13)

### Week 1 - May 4, 2026

- [x] **Artifact**: Start the pay syncing project and define the system shape.
- **Work connection**: Notice where real payment workflows depend on external systems.
- **Done when**: The project has a clear problem statement and first architecture notes.

### Week 2 - May 11, 2026

- [x] **Artifact**: Build out PayerSync docs and module boundaries.
- **Work connection**: Ask what owns state in a real payment flow.
- **Done when**: The main modules and responsibilities are documented.

### Week 3 - May 18, 2026

- [x] **Artifact**: Document reconciliation, idempotency, and failure modes in PayerSync.
- **Work connection**: Look for a real example where internal state and external state can disagree.
- **Done when**: The project explains how duplicate, delayed, or mismatched records are handled.

### Week 4 - May 25, 2026

- [x] **Artifact**: Finish the pay syncing project.
- **Work connection**: Pull out the strongest lessons for work design reviews.
- **Done when**: PayerSync is complete enough to use as a portfolio/proof-of-learning artifact.

### Week 5 - Jun 1, 2026

- [x] **Artifact**: Write a short PayerSync retrospective.
- **Work connection**: Identify which lessons transfer to Stripe, reconciliation, ledger, or payment-state work.
- **Done when**: You know what PayerSync taught you and what it did not teach deeply enough.

### Week 6 - Jun 8, 2026

- [x] **Artifact**: Complete and document the payment propensity ML project from June 9-10.
- **Work connection**: Think through how a payment-likelihood score could be misused in a real product.
- **Done when**: The project has a short README/case study covering goal, model, data, output, risk, and production monitoring.

### Week 7 - Jun 15, 2026

- [ ] **Artifact**: Rewrite the 52-week plan around payments fundamentals.
- **Work connection**: Ask your manager which payments area would be most useful for you to get deeper in first.
- **Done when**: The plan names the fundamentals, completed work, next project, and next 6 weeks.

### Week 8 - Jun 22, 2026

- [ ] **Artifact**: Create a payments vocabulary doc.
- **Work connection**: Collect the terms your team uses for payment state, ledger state, processor state, refunds, disputes, and payouts.
- **Done when**: You have plain-English definitions for at least 40 terms and notes on which ones your company uses differently.

### Week 9 - Jun 29, 2026

- [ ] **Artifact**: Draw a payment lifecycle state machine.
- **Work connection**: Compare it to a real payment flow at work.
- **Done when**: The diagram covers creation, auth, capture, settlement, refund, dispute, failure, and terminal states.

### Week 10 - Jul 6, 2026

- [ ] **Artifact**: Write a one-page guide to double-entry bookkeeping for engineers.
- **Work connection**: Find one internal ledger or reporting concept and map it to accounts, entries, and balances.
- **Done when**: You can explain debits, credits, balanced transactions, reversals, and derived balances without hand-waving.

### Week 11 - Jul 13, 2026

- [ ] **Artifact**: Scope the ledger project.
- **Work connection**: Ask what ledger-like record your team trusts when payment systems disagree.
- **Done when**: The README has goals, non-goals, account types, transaction types, invariants, and example flows.

### Week 12 - Jul 20, 2026

- [ ] **Artifact**: Design the ledger data model.
- **Work connection**: Compare it to how real payment records link to processor IDs, bank IDs, customer IDs, and internal IDs.
- **Done when**: Accounts, transactions, entries, metadata, idempotency keys, and balance queries are modeled.

### Week 13 - Jul 27, 2026

- [ ] **Artifact**: Review the ledger project design with your manager or a payments engineer.
- **Work connection**: Ask which design choice is most unrealistic compared with production payment systems.
- **Done when**: You have feedback, open questions, and a short implementation plan.

---

## Q2: Build the Ledger and Learn Money Accuracy (Weeks 14-26)

### Week 14 - Aug 3, 2026

- [ ] **Artifact**: Implement accounts, transactions, and ledger entries.
- **Work connection**: Notice where work systems store immutable facts versus mutable status.
- **Done when**: A balanced transaction can be created and read back.

### Week 15 - Aug 10, 2026

- [ ] **Artifact**: Add balance queries.
- **Work connection**: Ask whether balances at work are stored, derived, cached, or recomputed.
- **Done when**: Account balances derive from entries and tests prove balanced transactions stay balanced.

### Week 16 - Aug 17, 2026

- [ ] **Artifact**: Add pending and posted entries.
- **Work connection**: Find a real state where the customer sees success before settlement is final.
- **Done when**: The project can show pending balance, posted balance, and available balance.

### Week 17 - Aug 24, 2026

- [ ] **Artifact**: Add idempotency keys.
- **Work connection**: Identify one real operation that must be safe to retry.
- **Done when**: Replaying the same transaction request does not create duplicate ledger movement.

### Week 18 - Aug 31, 2026

- [ ] **Artifact**: Model a customer payment and processor fee.
- **Work connection**: Learn where fees appear in reports and who consumes that data.
- **Done when**: A payment can create entries for customer receivable/cash, processor fee, and net settlement.

### Week 19 - Sep 7, 2026

- [ ] **Artifact**: Model refunds and reversals.
- **Work connection**: Compare refund timing to original payment timing at work.
- **Done when**: Refunds create new entries instead of editing old entries.

### Week 20 - Sep 14, 2026

- [ ] **Artifact**: Model disputes and chargebacks.
- **Work connection**: Learn who owns dispute workflows and what state changes are visible to customers.
- **Done when**: The ledger can represent chargeback opened, funds withdrawn, chargeback won, and chargeback lost.

### Week 21 - Sep 21, 2026

- [ ] **Artifact**: Model payouts.
- **Work connection**: Trace one payment from customer payment to payout or settlement reporting.
- **Done when**: The project can explain gross amount, fees, net amount, and payout timing.

### Week 22 - Sep 28, 2026

- [ ] **Artifact**: Add metadata and external IDs.
- **Work connection**: Identify the IDs needed to debug a payment across internal DB, processor, bank, and customer support views.
- **Done when**: Entries can be traced by internal transaction ID, processor payment ID, payout ID, and customer/account ID.

### Week 23 - Oct 5, 2026

- [ ] **Artifact**: Add a reconciliation import format.
- **Work connection**: Ask what external reports your team uses to confirm money movement.
- **Done when**: The project can load a fake processor settlement report.

### Week 24 - Oct 12, 2026

- [ ] **Artifact**: Build reconciliation matching.
- **Work connection**: Study one real mismatch category: missing payment, fee mismatch, duplicate, timing lag, or stale status.
- **Done when**: The project reports matched, missing, duplicated, delayed, and amount-mismatched records.

### Week 25 - Oct 19, 2026

- [ ] **Artifact**: Add an exception queue.
- **Work connection**: Learn how operations or support resolves payment exceptions today.
- **Done when**: Mismatches can be assigned a reason, owner, status, and audit note.

### Week 26 - Oct 26, 2026

- [ ] **Artifact**: Write a Q2 case study on the ledger project.
- **Work connection**: Share the case study with your manager and ask what maps best to Staff-level expectations.
- **Done when**: You have a concise writeup explaining the ledger, tradeoffs, tests, and what you learned.

---

## Q3: Build Reliable Payment Workflows (Weeks 27-39)

### Week 27 - Nov 2, 2026

- [ ] **Artifact**: Choose the second project: card lifecycle simulator, ACH return simulator, or webhook replay tool.
- **Work connection**: Pick the one closest to work problems you actually see.
- **Done when**: The README has a problem statement, scope, and success criteria.

### Week 28 - Nov 9, 2026

- [ ] **Artifact**: Design the payment state machine.
- **Work connection**: Compare each state to real work states and identify overloaded or ambiguous states.
- **Done when**: States, transitions, terminal states, and invalid transitions are documented.

### Week 29 - Nov 16, 2026

- [ ] **Artifact**: Implement the happy path.
- **Work connection**: Trace a real happy path from request through processor/bank confirmation.
- **Done when**: One payment can move through the main lifecycle with clear state changes.

### Week 30 - Nov 23, 2026

- [ ] **Artifact**: Add webhook ingestion with dedupe.
- **Work connection**: Study how your team handles duplicate processor events.
- **Done when**: Duplicate events are ignored safely and logged.

### Week 31 - Nov 30, 2026

- [ ] **Artifact**: Add out-of-order event handling.
- **Work connection**: Find a real example where async events can arrive before synchronous processing finishes.
- **Done when**: The system can tolerate event ordering without corrupting state.

### Week 32 - Dec 7, 2026

- [ ] **Artifact**: Add retry behavior.
- **Work connection**: Identify which retries at work are safe and which need operator review.
- **Done when**: Transient failures retry, permanent failures stop, and unsafe retries require manual action.

### Week 33 - Dec 14, 2026

- [ ] **Artifact**: Add processor timeout handling.
- **Work connection**: Ask what happens if a request times out after the processor received it.
- **Done when**: The project can recover by checking external state before retrying money movement.

### Week 34 - Dec 21, 2026

- [ ] **Artifact**: Connect the workflow to the ledger project.
- **Work connection**: Map real payment state changes to ledger entries or reporting records.
- **Done when**: Payment lifecycle events produce balanced ledger entries where appropriate.

### Week 35 - Dec 28, 2026

- [ ] **Artifact**: Add an audit log.
- **Work connection**: Look at one real incident or escalation and ask what audit trail would make it easier.
- **Done when**: Major state changes record who/what caused them and what external IDs were involved.

### Week 36 - Jan 4, 2027

- [ ] **Artifact**: Add an operations runbook.
- **Work connection**: Compare it to how on-call or support would actually recover stuck payments.
- **Done when**: The runbook explains replay, retry, manual resolution, and escalation.

### Week 37 - Jan 11, 2027

- [ ] **Artifact**: Add observability metrics.
- **Work connection**: Identify metrics for stuck payments, decline rate, retry rate, reconciliation exceptions, and money at risk.
- **Done when**: The project exposes or documents the metrics that would make production drift visible.

### Week 38 - Jan 18, 2027

- [ ] **Artifact**: Write a failure-mode catalog.
- **Work connection**: Use the catalog to ask sharper questions in a work design review.
- **Done when**: The catalog covers timeout, duplicate, out-of-order, partial success, mismatch, reversal, and manual recovery.

### Week 39 - Jan 25, 2027

- [ ] **Artifact**: Present the project and payments lessons to your manager.
- **Work connection**: Ask which parts would be valuable to share with the team.
- **Done when**: You have feedback and a clear Q4 focus.

---

## Q4: Turn Payments Knowledge Into Staff-Level Evidence (Weeks 40-52)

### Week 40 - Feb 1, 2027

- [ ] **Artifact**: Write "Payments Fundamentals for Engineers."
- **Work connection**: Make it useful for engineers joining payment projects at work.
- **Done when**: The doc explains money movement, lifecycle state, ledger, reconciliation, reliability, risk, and business mechanics.

### Week 41 - Feb 8, 2027

- [ ] **Artifact**: Write a "State of Payments Knowledge" self-review.
- **Work connection**: Compare your current knowledge to what your manager expects from Staff-level domain ownership.
- **Done when**: You know what is strong, what is fuzzy, and what needs production exposure.

### Week 42 - Feb 15, 2027

- [ ] **Artifact**: Audit one real payment flow at work.
- **Work connection**: Use the fundamentals checklist against a real system.
- **Done when**: You have a private notes doc with states, IDs, external systems, risks, and open questions.

### Week 43 - Feb 22, 2027

- [ ] **Artifact**: Create a payments review checklist.
- **Work connection**: Use it in a TPD or design review.
- **Done when**: The checklist helps identify at least one real risk, ambiguity, or missing recovery path.

### Week 44 - Mar 1, 2027

- [ ] **Artifact**: Identify one practical improvement to propose at work.
- **Work connection**: Choose something small enough to be credible and useful.
- **Done when**: You have a problem statement, scope, affected teams, and expected impact.

### Week 45 - Mar 8, 2027

- [ ] **Artifact**: Draft a design note for the improvement.
- **Work connection**: Get early feedback from your manager or a payments engineer.
- **Done when**: The note explains current behavior, proposed change, risks, rollout, and measurement.

### Week 46 - Mar 15, 2027

- [ ] **Artifact**: Refine the design note with feedback.
- **Work connection**: Talk to one stakeholder outside your immediate team if the improvement crosses boundaries.
- **Done when**: The proposal is grounded in real constraints instead of theory alone.

### Week 47 - Mar 22, 2027

- [ ] **Artifact**: Ship or help ship one small payment reliability, observability, docs, or test improvement.
- **Work connection**: Make the improvement visible to your manager.
- **Done when**: There is a concrete artifact: PR, doc, dashboard, alert, runbook, or test.

### Week 48 - Mar 29, 2027

- [ ] **Artifact**: Write a short case study of the improvement.
- **Work connection**: Tie it to money risk, customer trust, operational load, or debugging speed.
- **Done when**: The case study states problem, change, tradeoff, and result.

### Week 49 - Apr 5, 2027

- [ ] **Artifact**: Map the year to Staff-level behaviors.
- **Work connection**: Ask your manager which evidence is strongest and which gaps remain.
- **Done when**: You have evidence for technical judgment, domain growth, writing, influence, and execution.

### Week 50 - Apr 12, 2027

- [ ] **Artifact**: Prepare a payments expertise portfolio index.
- **Work connection**: Organize it so your manager can quickly see the story.
- **Done when**: The index links PayerSync, ML case study, ledger project, workflow project, work artifacts, and retrospectives.

### Week 51 - Apr 19, 2027

- [ ] **Artifact**: Polish the strongest artifacts.
- **Work connection**: Remove vague language and make tradeoffs, risks, and outcomes concrete.
- **Done when**: The artifacts read like clear engineering work, not generic career content.

### Week 52 - Apr 26, 2027

- [ ] **Artifact**: Write the year-end retrospective and next-year plan.
- **Work connection**: Use it for the next manager growth conversation.
- **Done when**: You can explain what you learned, what you shipped, what changed at work, and what payments area you should go deeper on next.

---

## Monthly Check-In

At the end of each month, answer:

1. What payment concept became clearer?
2. What artifact did I create?
3. What did I connect to real work?
4. What did I ask my manager or a payments engineer?
5. What did I ship, document, or test?
6. What still feels fuzzy?

---

## Payments Review Checklist

Use this for toy projects and real design reviews:

- What is the source of truth?
- What external system can disagree with us?
- What ID links our record to the processor, bank, network, or ledger?
- What states are customer-visible?
- What states are internal-only?
- What states are terminal?
- What can run twice safely?
- What happens after timeout?
- What happens after partial success?
- How do we reconcile this later?
- What audit trail proves what happened?
- What manual recovery path exists?
- Who owns the financial loss if this fails?
- What sensitive data are we touching?
- What metric tells us this is healthy?

---

## Key Principles

**Build around money correctness**: In payments, accepting a request is the start. The harder work is knowing what happened, proving it later, and fixing it safely when systems disagree.

**Double-entry bookkeeping is worth learning deeply**: A ledger project will make debits, credits, balances, reversals, and reconciliation concrete.

**Use toy projects to ask better real questions**: The goal is not to build a fake fintech company. The goal is to make your work questions sharper.

**Write like an engineer**: Plain docs with examples, states, failure modes, and tradeoffs are better than polished vague summaries.

**Connect every project to work**: Each artifact should eventually help you in a design review, debugging session, incident discussion, or manager conversation.

# Reconciliation

Finding out whether your system agrees with external reality.

## Core Concepts

### What Reconciliation Is

Reconciliation is comparing two sets of records and explaining the differences.

- Your database says you processed $10,000 in payments today
- The processor's settlement report says $9,850
- Where's the $150?

That's reconciliation. Finding the answer.

### Why It Matters

Your internal records are not the source of truth for money. Banks and processors are. If your records don't match theirs, one of you is wrong. Usually it's you.

Reconciliation catches:
- Missing transactions
- Duplicate transactions
- Incorrect amounts
- Timing differences
- Fees you didn't account for
- Chargebacks and disputes

### Types of Reconciliation

**Transaction-level**: Match each transaction in your system to a transaction in the external report.

**Balance-level**: Compare aggregate balances. Faster but less precise. Good for sanity checks.

**Three-way**: Match your records, the processor's records, and the bank's records. More complex but catches more issues.

### Expected vs Unexpected Differences

Some differences are normal:

- **Timing**: You captured a payment at 11:59 PM. It settles tomorrow. Today's report won't include it.
- **Fees**: Processor takes 2.9% + $0.30. Your gross is $100, net is $97.01.
- **Currency conversion**: You charged €100, settled in USD at today's rate.

Unexpected differences need investigation:

- Transaction in your system but not in settlement report
- Transaction in settlement report but not in your system
- Amount mismatch that isn't explained by fees

### Exception Handling

When reconciliation finds a mismatch:

1. **Log it** - Record the discrepancy with enough detail to investigate
2. **Categorize it** - Is this a known pattern or something new?
3. **Route it** - Some exceptions auto-resolve, some need human review
4. **Resolve it** - Fix the root cause, not just the symptom
5. **Close it** - Document what happened and why

## What Can Go Wrong

- **Silent failures** - Reconciliation runs but nobody looks at the exceptions
- **Too many exceptions** - Team ignores them because there are always hundreds
- **Missing reports** - Settlement report didn't arrive, reconciliation didn't run
- **Stale data** - Reconciling against yesterday's data, missing today's problems
- **Manual corrections** - Someone "fixes" the data without understanding why it was wrong

## How This Connects to My Projects

**payer-sync**: The reconciler matches ERA (remittance) records to VCC (card) records. When they match, the payment can be processed. When they don't, it goes to an exception queue.

**ledger-db**: The ledger is the internal source of truth. Reconciliation would compare ledger balances to external settlement reports.

## Questions to Ask About Real Systems

- What external reports do we reconcile against?
- How often does reconciliation run?
- What's the typical exception rate?
- Who reviews exceptions?
- How long do exceptions stay open?
- What's the process for correcting mismatches?

## Keywords

Reconciliation, settlement report, exception queue, timing difference, three-way reconciliation, balance reconciliation, transaction matching, discrepancy

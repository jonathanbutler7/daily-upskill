# Ledger & Accounting

Double-entry bookkeeping: every financial event creates entries that balance.

## Core Concepts

### Double-Entry Bookkeeping

Every transaction creates at least two entries. The sum of debits equals the sum of credits. Always.

This isn't just accounting tradition. It's a correctness invariant. If debits ≠ credits, something is wrong.

### Debits and Credits

Debits and credits don't mean "money in" and "money out." Their meaning depends on the account type:

| Account Type | Debit | Credit |
|--------------|-------|--------|
| Asset | Increase | Decrease |
| Liability | Decrease | Increase |
| Revenue | Decrease | Increase |
| Expense | Increase | Decrease |

Example: Customer pays $100.

- Debit Cash (asset) $100 - Cash increases
- Credit Revenue $100 - Revenue increases

Both entries are positive events. The words "debit" and "credit" just indicate which side of the equation.

### Account Types

- **Assets** - What you own (cash, receivables, inventory)
- **Liabilities** - What you owe (payables, customer deposits, reserves)
- **Revenue** - Money earned
- **Expenses** - Money spent
- **Contra accounts** - Offset another account (e.g., refunds contra revenue)

### Derived Balances

Account balances are derived from entries, not stored directly. To get a balance:

```sql
SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END)
FROM entries
WHERE account_id = ?
```

This means you can always reconstruct balances from the entry history.

### Immutability

Never edit or delete entries. If something is wrong, add a reversal entry.

Wrong: `UPDATE entries SET amount = 90 WHERE id = 123`

Right: Add a new entry that reverses the original, then add the correct entry.

This preserves audit history. You can always explain what happened and when.

### Pending vs Posted

- **Pending** - Entry is recorded but not finalized (authorization, hold)
- **Posted** - Entry is finalized and affects the balance

Pending entries might not affect the "available balance" but do affect the "pending balance."

## What Can Go Wrong

- **Unbalanced transactions** - Debits ≠ credits. Invariant violated.
- **Edited history** - Someone updated an entry instead of reversing it. Audit trail broken.
- **Missing entries** - Event happened but wasn't recorded. Balances are wrong.
- **Double-posting** - Same event recorded twice. Balances are wrong.
- **Pending stuck** - Entry never transitions from pending to posted.

## How This Connects to My Projects

**ledger-db**: Implemented double-entry with Postgres. `post_transfer` function creates balanced entries atomically. Idempotency key prevents double-posting. Insufficient funds check happens before any writes.

**payer-sync**: The reconciliation step compares internal ledger state to external settlement reports. Mismatches indicate missing or incorrect entries.

## Questions to Ask About Real Systems

- How do we ensure transactions always balance?
- How do we handle corrections? (Reversals vs edits)
- What's the difference between pending and posted?
- How do we calculate available balance vs total balance?
- How do we audit entry history?

## Keywords

Double-entry bookkeeping, debit, credit, ledger, journal entry, account balance, immutability, reversal, pending, posted, audit trail, contra account

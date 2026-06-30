# Project Log

This started as a daily time tracker for 15-minute exercises. That format didn't last. Now it's a log of projects and what they taught me.

## Projects

### payer-sync (May 2026)

Healthcare payment reconciliation system. Ingests ERA and VCC files, reconciles them, processes payments, writes back to PMS.

**What I learned**:
- System boundary thinking (what Postgres owns vs what the app owns)
- Reconciliation and source-of-truth questions
- Idempotency and retry design
- State machines for payment lifecycle
- Failure-mode analysis
- Technical writing and diagrams

**Status**: Done around May 27, 2026.

### ledger-db (May-June 2026)

Double-entry ledger implementation in Postgres. Accounts, transactions, entries, balances.

**What I learned**:
- Double-entry bookkeeping (debits, credits, account types)
- Postgres functions for atomic operations
- Idempotency via unique constraints
- Immutability (correct with reversals, don't edit history)

**Status**: Core functionality complete. Can extend with more scenarios.

### will-they-pay (June 2026)

ML model predicting payment likelihood. Built for Weave hackathon.

**What I learned**:
- Feature engineering for payment behavior
- Model evaluation and selection
- Connecting ML predictions to business decisions
- Risks: bias, overconfidence, stale data

**Status**: Done June 9-10, 2026. Document as case study, not career direction.

---

## Original Exercise Log

These are the daily exercises from the original "15-minute habit" phase.

| Date       | Area              | Exercise File                          | Notes |
|------------|-------------------|----------------------------------------|-------|
| 2026-03-21 | -                 | -                                      | Repo initialized |
| 2026-03-21 | Python Foundations | exercises/data_transformations_260321.py | List comprehensions, f-strings (8 mins) |
| 2026-03-23 | Python Foundations | exercises/typing_and_pydantic_260323.py | Pydantic models, Field constraints (15 mins) |
| 2026-03-24 | Python Foundations | exercises/context_managers_260324.py | Context managers, DictReader (11 mins) |
| 2026-03-25 | Python Foundations | exercises/functional_patterns_260325.py | map, filter, reduce (15 mins) |
| 2026-03-27 | Python Foundations | exercises/decorators_260327.py | Retry decorator (30 mins) |
| 2026-03-30 | Python Foundations | exercises/testing_mocks260330.py | pytest fixtures, MagicMock (45 mins) |
| 2026-04-02 | Python Foundations | module_1_quiz.ipynb | Quiz (20 mins) |
| 2026-04-03 | Python Foundations | module_1_quiz.ipynb | Quiz (25 mins) |
| 2026-04-04 | Python Foundations | module_1_quiz.ipynb | Quiz (30 mins) |
| 2026-05-05 | System Design | design patterns cheat sheet | Memorizing tradeoffs (30 mins) |
| 2026-05-06 | System Design | design patterns cheat sheet | All subcategories memorized (15 mins) |
| 2026-05-06 | System Design | design-patterns-reference.md | Read through (15 mins) |

This format stopped being useful once I shifted to project-based learning.

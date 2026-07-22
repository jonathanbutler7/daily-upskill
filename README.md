# Daily Upskill

A workspace for building payments expertise and Staff-level technical authority.

## What This Is

This repo started as a "15-minute daily practice" experiment for Python and ML. That didn't stick. What actually worked was building real projects that forced me to learn concepts deeply.

Small exercises felt pointless. Small investment, small return. Larger projects with real complexity taught me more and felt worth the time.

The repo has evolved into three things:

1. **Projects** - Working systems that implement payments concepts
2. **Knowledge** - Reference docs for payments, system design, and career planning
3. **Learning** - The original exercise structure, mostly archived

## Structure

```
daily-upskill/
├── projects/                      # Working code
│   ├── payer-sync/               # Healthcare payment reconciliation
│   ├── ledger-db/                # Double-entry ledger in Postgres
│   ├── ideas/                    # Future payments product designs
│   │   ├── wallet/               # Stored-value wallet product design
│   │   ├── reconciliation-engine/ # Reconciles reports against ledger state
│   │   └── rail-sim/             # Payment rail simulator design
│   └── will-they-pay/            # ML payment propensity model
│
├── knowledge/                     # Reference documentation
│   ├── payments-fundamentals/    # Core payments concepts
│   ├── system-fundamentals/      # Production infrastructure
│   ├── system-design/            # Design patterns
│   └── career-path/              # Staff → Director roadmap
│
├── learning/                      # Original learning materials
│   ├── exercises/                # Python exercises (mostly done)
│   ├── ROADMAP.md                # Learning progression
│   └── SKILLS_TRACKER.md         # Project log
│
└── archive/                       # Deprecated content
    └── old-stuff/                # Early experiments
```

## Current Focus

Building payments domain expertise for Staff Engineer promotion. The goal is to become the engineer people pull into design conversations when money movement, correctness, or risk is involved.

The current payments architecture work is split across separate products:

- `ledger-db` owns durable double-entry accounting, balances, idempotency, reversals, and the ledger state needed by downstream systems.
- `projects/ideas/reconciliation-engine` owns the design for external report ingestion, matching, exceptions, and reconciliation resolution workflow.
- `projects/ideas/wallet` and `projects/ideas/rail-sim` sit on either side of the ledger: wallet as the user-facing stored-value product, and rail-sim as a future external payment rail simulator.

See `knowledge/career-path/` for the full plan.

## How I Learn

- **Projects over exercises** - Building systems teaches more than isolated practice
- **Concrete first** - See the problem before the theory
- **Write it down** - Documenting forces clarity
- **Connect to work** - Learning sticks when it applies to real systems

---

**Started**: March 2026  
**Current track**: Payments fundamentals, ledger correctness, reconciliation systems, reliability

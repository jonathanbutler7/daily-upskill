# Learning Roadmap

**Goal**: Build Staff-level technical authority in payments  
**Timeline**: May 2026 - May 2027  
**Current date**: June 30, 2026

## How This Plan Changed

The original plan was a 5-module Python/ML curriculum:
1. Python Foundations
2. ML Engineering & Tools
3. Data Engineering & Architecture
4. System Design & Backend Patterns
5. Capstone Project

Module 1 worked. The rest didn't happen because I realized:
- ML isn't my career direction
- Generic skill-building doesn't lead to Staff
- I needed depth in one domain, not breadth across many

The new plan focuses on payments. It's the domain I work in, it's complex enough to build real expertise, and it's where I can demonstrate Staff-level impact.

---

## Phase 1: Foundations (Complete)

- [x] Python fundamentals (Module 1 exercises)
- [x] First payments project (`payer-sync`)
- [x] Ledger basics (`ledger-db`)
- [x] ML payment propensity model (`will-they-pay`)

These projects taught me more than the original curriculum would have. They're now evidence for Staff promotion.

---

## Phase 2: Payments Depth (Current)

### Payments Fundamentals

Document the core concepts in `knowledge/payments-fundamentals/`:

- [x] Money movement (authorization, capture, settlement, funding)
- [x] Payment state machines and lifecycle
- [x] Double-entry ledger and accounting
- [x] Reconciliation and exception handling
- [x] Reliability patterns (idempotency, retries, webhooks)
- [x] Risk, fraud, and compliance basics

### Projects

Build systems that force deeper understanding:

- [ ] Payment state machine with webhook handling
- [ ] Reconciliation engine for Stripe settlements
- [ ] Idempotent payment API with retry logic

### Work Connection

Apply learning to real Weave systems:

- [ ] Document Weave's payment architecture
- [ ] Map payment flows end-to-end
- [ ] Identify reliability gaps and propose improvements

---

## Phase 3: Staff Evidence (Q4 2026)

- [ ] Write design docs for payment features
- [ ] Create payment reliability checklist
- [ ] Build portfolio of payment artifacts
- [ ] Prepare Staff promotion packet

---

## Original Modules (Archived)

The original Python/ML modules are preserved in `archive/old-stuff/` for reference. They're not part of the current plan.

| Module | Status |
|--------|--------|
| Module 1: Python Foundations | ✅ Complete |
| Module 2: ML Engineering & Tools | ❌ Abandoned |
| Module 3: Data Engineering & Architecture | ❌ Abandoned |
| Module 4: System Design & Backend Patterns | ❌ Abandoned |
| Module 5: Capstone Project | ❌ Abandoned |

The concepts from these modules (Pydantic, decorators, testing) were useful. The curriculum structure wasn't.

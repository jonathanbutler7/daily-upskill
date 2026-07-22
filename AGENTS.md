# Agent Response Guidelines

## Voice And Writing Style

Do not make my writing sound like AI-generated technical content.

When editing, reviewing, or suggesting prose:
- Prefer plain human sentences over polished-sounding abstractions.
- Keep my natural voice unless I explicitly ask for a more formal tone.
- Avoid inflated consulting/product-strategy phrasing.
- Preserve useful technical precision, but make it sound direct and grounded.
- Be concise. If it can be said shorter, say it shorter.
- Avoid the "not X, but Y" explanation pattern.
- Do not start a sentence or paragraph by saying what something is not before saying what it is. Lead with the direct claim in plain language.
- Avoid sentences that set up and negate expectations, such as "X isn't just about Y."
- Use contrast only when the distinction is genuinely necessary for correctness.

Bad:
- The goal is not to build a real ACH gateway. The goal is to create a small system that makes delayed settlement easy to understand.
- Real payment systems do not just post money immediately. They deal with files, windows, settlement, returns, and reconciliation.

Better:
- `rail-sim` is a small simulator for delayed settlement, returns, reconciliation, and external rail state.
- Real payment systems handle files, processing windows, settlement dates, returns, reversals, duplicate references, and reconciliation reports.

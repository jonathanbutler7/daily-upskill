# Concurrency

1. Can multiple reconciler workers run in parallel?
2. If yes, how do you prevent race conditions when matching the same ERA/VCC pair?
3. What DB primitives (row locks, optimistic versioning, advisory locks) enforce single-processing guarantees?
4. How does this interact with your polling model—does each poll cycle claim a batch of work?

This is the most critical gap. The PRD has out-of-order arrival, 5-day waiting windows, and reprocessing on every new file. Without a clear concurrency model, you'll either have race conditions or unnecessary serialization bottlenecks.


# CONCEPT: Async/Await for Concurrent I/O

> **Related Exercise**: `exercises/async_await_260414.py`

## When to Use

When you need to perform many I/O operations concurrently (API calls, database queries, file reads) without the overhead of threads. Async is ideal for high-concurrency I/O-bound workloads.

## Core Concept

`async`/`await` is Python's syntax for writing asynchronous code. An `async def` function returns a **coroutine** that doesn't run until awaited. The **event loop** manages switching between coroutines when they hit `await` points (I/O operations). Unlike threading, async uses a single thread—no GIL contention, no race conditions from shared state.

## Pattern

```python
import asyncio
import time

async def fetch_data(endpoint_id: int, delay: float) -> str:
    """Simulates an API call with network latency"""
    print(f"[{time.time():.2f}] Starting fetch from endpoint {endpoint_id}")
    await asyncio.sleep(delay)  # Non-blocking wait
    print(f"[{time.time():.2f}] Completed fetch from endpoint {endpoint_id}")
    return f"Data from endpoint {endpoint_id}"

async def main():
    # Sequential (slow)
    start = time.time()
    result1 = await fetch_data(1, 1)
    result2 = await fetch_data(2, 1)
    result3 = await fetch_data(3, 1)
    print(f"Sequential: {time.time() - start:.2f}s")  # ~3s

    # Concurrent (fast!)
    start = time.time()
    results = await asyncio.gather(
        fetch_data(1, 1),
        fetch_data(2, 1),
        fetch_data(3, 1)
    )
    print(f"Concurrent: {time.time() - start:.2f}s")  # ~1s
    print(f"Results: {results}")

asyncio.run(main())
```

## Real-World DE/ML Scenario

- **Batch API calls**: Fetching embeddings from OpenAI for 100 documents concurrently
- **Database queries**: Running multiple independent queries in parallel
- **Web scraping**: Downloading pages from multiple URLs simultaneously
- **Microservices**: Aggregating data from multiple backend services

## Comparison with TypeScript

| Python | TypeScript/Node.js |
|--------|-------------------|
| `async def` | `async function` |
| `await` | `await` |
| `asyncio.gather()` | `Promise.all()` |
| `asyncio.run(main())` | Top-level await or `.then()` |
| `asyncio.sleep(1)` | `new Promise(r => setTimeout(r, 1000))` |

## Key Differences: Async vs Threading

| Aspect | `asyncio` | `threading` |
|--------|-----------|-------------|
| Concurrency model | Single-threaded, cooperative | Multi-threaded, preemptive |
| GIL impact | No GIL contention | GIL limits CPU parallelism |
| Best for | High-concurrency I/O | Moderate I/O, legacy blocking code |
| Complexity | Requires async-compatible libraries | Works with any blocking code |

## Keywords to Search

- "Python asyncio tutorial"
- "asyncio.gather vs asyncio.wait"
- "Python async vs threading"
- "aiohttp async HTTP client"

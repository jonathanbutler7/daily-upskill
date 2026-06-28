# CONCEPT: Understanding the GIL and Threading

> **Related Exercise**: `exercises/gil_threading_260414.py`

## When to Use

When you need to understand why multi-threaded Python code doesn't always run faster, and when to choose threading vs multiprocessing for concurrent workloads.

## Core Concept

The **Global Interpreter Lock (GIL)** is a mutex that protects access to Python objects, preventing multiple threads from executing Python bytecode simultaneously. This means CPU-bound tasks don't benefit from threading—only one thread runs at a time. However, I/O-bound tasks (network calls, file reads) release the GIL while waiting, so threading *does* help there.

## Pattern

```python
import threading
import time

def cpu_bound_task(n):
    """Simulates CPU-intensive work (GIL prevents parallel execution)"""
    count = 0
    for i in range(n):
        count += i ** 2
    return count

def io_bound_task(delay):
    """Simulates I/O work (GIL is released during sleep)"""
    time.sleep(delay)
    return f"Completed after {delay}s"

# Single-threaded CPU-bound
start = time.time()
cpu_bound_task(10_000_000)
cpu_bound_task(10_000_000)
print(f"Single-threaded CPU: {time.time() - start:.2f}s")

# Multi-threaded CPU-bound (GIL prevents speedup!)
start = time.time()
threads = [
    threading.Thread(target=cpu_bound_task, args=(10_000_000,)),
    threading.Thread(target=cpu_bound_task, args=(10_000_000,))
]
for t in threads:
    t.start()
for t in threads:
    t.join()
print(f"Multi-threaded CPU: {time.time() - start:.2f}s")

# Multi-threaded I/O-bound (GIL released, threading helps!)
start = time.time()
threads = [
    threading.Thread(target=io_bound_task, args=(2,)),
    threading.Thread(target=io_bound_task, args=(2,))
]
for t in threads:
    t.start()
for t in threads:
    t.join()
print(f"Multi-threaded I/O: {time.time() - start:.2f}s")  # ~2s, not 4s!
```

## Real-World DE/ML Scenario

- **CPU-bound**: Model inference, data transformations, numerical computations → Use `multiprocessing` to bypass GIL
- **I/O-bound**: API calls, database queries, file downloads → Use `threading` or `asyncio`
- **Mixed workloads**: Use a thread pool for I/O and process pool for CPU work

## Comparison with TypeScript

| Python | TypeScript/Node.js |
|--------|-------------------|
| GIL limits CPU parallelism | Single-threaded event loop (similar limitation) |
| `threading` for I/O concurrency | `Promise.all()` for I/O concurrency |
| `multiprocessing` for CPU parallelism | Worker threads for CPU parallelism |
| `concurrent.futures.ThreadPoolExecutor` | No direct equivalent (use worker threads) |

## Keywords to Search

- "Python GIL explained"
- "Python threading vs multiprocessing"
- "When does Python release the GIL"
- "concurrent.futures ThreadPoolExecutor vs ProcessPoolExecutor"

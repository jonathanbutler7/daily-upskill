# SKILL: Functional patterns for data pipelines (map, filter, reduce)

**When to use**: Building data cleaning or transformation pipelines where you want to chain operations together in a clear, functional way.

**Core Concept**: 
Higher-order functions that take a function as an argument and apply it to an iterable.
- `map(func, iterable)`: Applies `func` to every item in `iterable`.
- `filter(func, iterable)`: Keeps items where `func(item)` is `True`.
- `reduce(func, iterable)`: Cumulatively applies `func` to items to reduce them to a single value (requires `from functools import reduce`).

**Pattern**:
```python
from functools import reduce

# Map: list(map(lambda x: x.strip().lower(), raw_strings))
# Filter: list(filter(lambda x: x > 0, numbers))
# Reduce: reduce(lambda acc, x: acc + x, numbers, 0)
```

**Real-World DE/ML Scenario**:
Processing a stream of raw logs to extract specific metrics.

```python
# Raw logs: logs = [" INFO: User login ", " ERROR: DB Timeout ", " INFO: User logout "]
# 1. Clean (map): ["INFO: User login", "ERROR: DB Timeout", "INFO: User logout"]
# 2. Filter (errors): ["ERROR: DB Timeout"]
# 3. Aggregate (count): 1
```

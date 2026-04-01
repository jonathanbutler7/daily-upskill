# CONCEPT: Writing Decorators for Reusable Logic

> **Related Exercise**: `exercises/decorators_260327.py`

## When to Use

Adding cross-cutting concerns (logging, metrics, retries, authentication) to functions without modifying their core logic.

## Core Concept

A decorator is a function that wraps another function to extend its behavior. Functions in Python are first-class objects, meaning they can be passed as arguments and returned from other functions. The `@decorator` syntax is shorthand for `func = decorator(func)`.

## Pattern

```python
import functools

def log_execution(func):
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        print(f"Calling {func.__name__}")
        result = func(*args, **kwargs)
        print(f"{func.__name__} returned {result}")
        return result
    return wrapper

@log_execution
def add(a, b):
    return a + b
```

## Real-World DE/ML Scenario

- **Logging**: Automatically log function inputs, outputs, or execution time.
- **Metrics**: Track how many times a function is called or its latency (e.g., Prometheus metrics).
- **Retries**: Automatically retry a function if a specific exception is raised (great for flaky APIs).

## Decorators with Arguments

When your decorator needs arguments (like `@retry(max_retries=3)`), you need an extra layer of nesting:

```python
def repeat(times):
    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            for _ in range(times):
                result = func(*args, **kwargs)
            return result
        return wrapper
    return decorator

@repeat(times=3)
def greet(name):
    print(f"Hello, {name}!")
```

## Keywords to Search

- "Python decorator with arguments"
- "functools.wraps purpose"
- "Python closure decorator"

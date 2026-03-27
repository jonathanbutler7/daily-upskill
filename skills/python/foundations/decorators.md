# Writing Decorators for Reusable Logic

In Python, a decorator is a powerful pattern that allows you to wrap functions to modify or extend their behavior cleanly.

## Core Concepts

### 1. Functions are First-Class Objects
Functions can be passed as arguments, assigned to variables, and returned by other functions.

### 2. Inner (Nested) Functions
A function defined inside another function. It has access to variables defined in the outer function's scope (closures).

### 3. Decorator Syntax (`@`)
Using `@decorator_name` is syntactic sugar for `my_function = decorator_name(my_function)`.

### 4. `*args` and `**kwargs`
These allow your wrapper function to accept and forward any number of positional and keyword arguments to the original function.

### 5. `functools.wraps`
This is critical. It's a decorator that you apply to your inner wrapper function. It copies the original function's name, docstring, and other metadata to the wrapper, making it "behave" like the original function when inspected (e.g., in debugging or documentation generation).

## Typical Use Cases
- **Logging**: Automatically log function inputs, outputs, or execution time.
- **Metrics**: Track how many times a function is called or its latency (e.g., Prometheus metrics).
- **Retries**: Automatically retry a function if a specific exception is raised (great for flaky APIs).
- **Authentication/Authorization**: Ensure a user has permission before executing a function.
- **Caching/Memoization**: Cache expensive function results based on their inputs.

## Pro-tip: Decorators with Arguments
When your decorator needs arguments (like `@retry(max_retries=3)`), you need an extra layer of nesting:
1. Outer function receives the decorator arguments.
2. Middle function receives the function to be decorated.
3. Inner function (the wrapper) receives the function's arguments.

### Example: A Logging Decorator
A simple decorator that logs when a function starts and ends.

```python
import functools

def log_execution(func):
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        print(f"DEBUG: Starting {func.__name__} with args: {args}, kwargs: {kwargs}")
        result = func(*args, **kwargs)
        print(f"DEBUG: {func.__name__} finished with result: {result}")
        return result
    return wrapper

@log_execution
def add(a, b):
    return a + b
```

### Example: Decorator with Arguments
When you need to pass parameters to the decorator itself (like a retry count or a timeout).

```python
import functools

def repeat(times):
    """Wait, what? A function that returns a decorator!"""
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

# CONCEPT: Using Protocols for Structural Typing (Go-style Interfaces)

> **Related Exercise**: `exercises/protocols_260327.py`

## When to Use

Defining interfaces that classes can implement implicitly, without explicit inheritance. This is useful for decoupling code and enabling duck typing with type safety.

## Core Concept

`typing.Protocol` (Python 3.8+) allows for **structural subtyping**. Instead of explicitly inheriting from a base class (nominal subtyping), a class is considered an implementation of a Protocol if it simply has the required methods and attributes. This is similar to how interfaces work in Go.

## Pattern

```python
from typing import Protocol, runtime_checkable

@runtime_checkable  # Allows isinstance(obj, Reader) at runtime
class Reader(Protocol):
    def read(self) -> str:
        ...  # Ellipsis is idiomatic for Protocol definitions

class FileReader:
    def read(self) -> str:
        return "Reading from a file..."

def process_data(reader: Reader):
    print(reader.read())

# Works because FileReader has a read() method!
process_data(FileReader())
```

## Real-World DE/ML Scenario

Defining a `DataSource` protocol that can be implemented by `S3Reader`, `LocalFileReader`, or `APIFetcher` without them needing to inherit from a common base class.

## Comparison with TypeScript (Optional)

| Python Protocol | TypeScript Interface |
|-----------------|---------------------|
| Implicit (structural) | Implicit (structural) |
| `@runtime_checkable` for `isinstance` | Built-in |

## Keywords to Search

- "Python Protocol vs ABC"
- "structural subtyping Python"
- "typing.Protocol runtime_checkable"

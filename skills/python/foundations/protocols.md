# Using Protocols for Structural Typing (Go-style Interfaces)

In Python, `typing.Protocol` (introduced in Python 3.8) allows for **structural subtyping**, which is very similar to how interfaces work in Go. Instead of explicitly inheriting from a base class (nominal subtyping), a class is considered an implementation of a Protocol if it simply has the required methods and attributes.

## Why Protocols?

- **Decoupling**: You don't need to import a base class to implement an interface.
- **Go-style**: If it walks like a duck and quacks like a duck, it's a duck.
- **Better Tooling**: IDEs and `mypy` can verify that your classes meet the requirements of the Protocol.

## Example

```python
from typing import Protocol, runtime_checkable

@runtime_checkable  # Allows using isinstance(obj, Reader) at runtime
class Reader(Protocol):
    def read(self) -> str:
        ...  # The ellipsis is idiomatic for Protocol definitions

class FileReader:
    def read(self) -> str:
        return "Reading from a file..."

class APIReader:
    def read(self) -> str:
        return "Reading from an API..."

def process_data(reader: Reader):
    print(reader.read())

# Works because both FileReader and APIReader have a read() method!
process_data(FileReader())
process_data(APIReader())
```

## Comparison with Go

| Feature | Python Protocol | Go Interface |
|---------|----------------|--------------|
| Implementation | Implicit (Structural) | Implicit (Structural) |
| Inheritance | Not required | Not required |
| Runtime check | Via `@runtime_checkable` | Implicit/Built-in |
| Multi-method | Supported | Supported |

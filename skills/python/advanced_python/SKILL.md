# SKILL: Advanced Python Comprehensions & Iterators

**When to use**: Anytime you need clean, fast data transformations.

**Pattern**:
```python
# Good
result = [x**2 for x in data if x > 0]

# Better (generator)
result = (x**2 for x in data if x > 0)

# With itertools
from itertools import islice, chain
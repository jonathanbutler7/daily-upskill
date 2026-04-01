# CONCEPT: Efficient Data Handling (Context Managers & File I/O)

> **Related Exercise**: `exercises/context_managers_260324.py`

## When to Use

Reading and writing large datasets from local or remote storage, ensuring system resources (like file handles) are properly released even if an error occurs.

## Core Concept

The `with` statement (Context Manager) simplifies resource management. Instead of manually opening and closing files, `with` handles the cleanup automatically. This is critical in Data Engineering when processing thousands of files or large CSVs where forgetting to close a file could lead to memory leaks or locked resources.

## Pattern

```python
# The 'with' statement ensures f.close() is called automatically
with open('data.csv', 'r') as f:
    for line in f:
        # Process line by line to keep memory usage low
        process(line)
```

## Real-World DE/ML Scenario

Processing a multi-gigabyte training dataset from a CSV file. Instead of loading the whole thing into memory (which could crash), you use a context manager to read it line by line or in chunks.

## Keywords to Search

- "Python context manager __enter__ __exit__"
- "contextlib contextmanager decorator"
- "Python with statement resource management"

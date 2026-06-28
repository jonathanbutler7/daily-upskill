# CONCEPT: Practical Data Transformations (Comprehensions & Generators)

> **Related Exercise**: `exercises/data_transformations_260321.py`

## When to Use

Cleaning, filtering, or transforming small to medium datasets in memory before passing them to a model or pipeline.

## Core Concept

Comprehensions are concise ways to create new lists, sets, or dictionaries. Generators are memory-efficient versions that only produce values as needed (perfect for large files).

## Pattern

```python
# List comprehension: [expression for item in iterable if condition]
# Dictionary comprehension: {key_expr: val_expr for item in iterable}
# Generator expression: (expression for item in iterable if condition)
```

## Real-World DE/ML Scenario

Imagine you have a list of raw sensor readings and need to filter out negative values while rounding the rest.

```python
# Raw data: sensor_data = [12.4, -1.0, 15.6, 22.1, -5.3]
# Transformation: rounded_positive = [round(x) for x in sensor_data if x >= 0]
```

## Keywords to Search

- "Python list comprehension vs generator"
- "Python dictionary comprehension"
- "memory efficient iteration Python"

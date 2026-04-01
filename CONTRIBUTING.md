# Contributing to Daily Upskill

This guide explains how to add new content to this repository consistently.

## Adding a New Concept

1. **Create the concept file** in `concepts/python/foundations/` (or the appropriate subdirectory for future modules).
2. **Use the template**: Copy `concepts/CONCEPT_TEMPLATE.md` and fill in each section.
3. **Keep it concise**: The user prefers practical examples over long documentation.

### Required Sections

| Section | Purpose |
|---------|---------|
| `> **Related Exercise**` | Link to the corresponding exercise file |
| `## When to Use` | One-liner for quick scanning |
| `## Core Concept` | 2-3 sentences max |
| `## Pattern` | Minimal, reusable code snippet |
| `## Real-World DE/ML Scenario` | Practical example |
| `## Keywords to Search` | Encourage self-research |

### Optional Sections

- `## Comparison with TypeScript` — Useful for bridging JS/TS knowledge.

---

## Adding a New Exercise

1. **File naming**: `exercises/taskname_YYMMDD.py` (e.g., `decorators_260327.py`).
2. **Add a header comment** linking back to the concept:
   ```python
   # Related Concept: concepts/python/foundations/decorators.md
   ```
3. **Provide a skeleton**: Imports, function signatures, and hints — but leave the core logic for the user to implement.
4. **Include test cases** or a `if __name__ == "__main__":` block for quick validation.

---

## Logging Progress

After completing an exercise, add a row to `SKILLS_TRACKER.md`:

```markdown
| 2026-03-27 | Python Foundations | exercises/decorators_260327.py | Built a retry decorator. (30 mins) |
```

---

## Committing Changes

Use clear, descriptive commit messages:

```
feat(concepts): add decorators concept with retry example
feat(exercises): add decorators exercise skeleton
docs: update SKILLS_TRACKER with decorators progress
```

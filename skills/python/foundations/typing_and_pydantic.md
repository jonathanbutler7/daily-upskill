# SKILL: Robust Data Validation (Typing & Pydantic)

**When to use**: Validating external data (APIs, JSON files, sensor inputs) before it enters your processing pipeline or ML model.

**Core Concept**: 
Python's type hints help catch bugs during development, but `Pydantic` enforces those types at runtime. If the data doesn't match the schema, Pydantic raises a clear error instead of letting a broken value crash your app later.

**Pattern**:
```python
from pydantic import BaseModel, Field

class SensorReading(BaseModel):
    id: int
    value: float = Field(..., gt=0) # Must be greater than 0
    unit: str = "Celsius"
```

**Real-World DE/ML Scenario**:
Ensuring that incoming training data has all required features with the correct numerical ranges before feeding it into a Scikit-Learn or PyTorch model.

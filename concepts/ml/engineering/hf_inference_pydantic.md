# CONCEPT: Standardizing Hugging Face Inference with Pydantic

> **Related Exercise**: `exercises/ml/hf_inference_pydantic_260405.py`

## When to Use

Wrapping Hugging Face pipeline outputs in typed schemas so downstream code (APIs, eval scripts, logging) always receives predictable data — not raw dicts.

## Core Concept

Hugging Face pipelines return plain Python dicts (or lists of dicts). The shape varies by task (`text-classification`, `ner`, `text-generation`, etc.). Using Pydantic BaseModels as the I/O contract means you validate the raw output once at the boundary and work with typed objects everywhere else.

## Pattern

```python
from pydantic import BaseModel, Field
from transformers import pipeline

class InferenceRequest(BaseModel):
    text: str
    model_id: str = "distilbert-base-uncased-finetuned-sst-2-english"

class SentimentResult(BaseModel):
    label: str          # e.g. "POSITIVE" | "NEGATIVE"
    score: float = Field(..., ge=0.0, le=1.0)

class InferenceResponse(BaseModel):
    request: InferenceRequest
    results: list[SentimentResult]

def run_inference(request: InferenceRequest) -> InferenceResponse:
    pipe = pipeline("text-classification", model=request.model_id)
    raw: list[dict] = pipe(request.text)
    return InferenceResponse(
        request=request,
        results=[SentimentResult(**item) for item in raw],
    )
```

## Real-World ML Engineering Scenario

A model-serving microservice accepts user text via an API. The handler parses the request into an `InferenceRequest`, calls the pipeline, validates the output into an `InferenceResponse`, then serializes it back to JSON — all with full type safety and no manual key lookups.

## Key Pydantic Tools Used Here

| Tool | Purpose |
|------|---------|
| `Field(ge=0.0, le=1.0)` | Score must be a probability |
| `model_dump()` | Serialize response to dict/JSON |
| `model_validate(raw_dict)` | Parse a raw dict into a model |

## Keywords to Search

- "Pydantic model_validate"
- "Hugging Face pipeline output format"
- "Pydantic Field ge le"
- "typing Literal Python"

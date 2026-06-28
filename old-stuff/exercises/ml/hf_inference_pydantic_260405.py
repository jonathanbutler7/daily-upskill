# Related Concept: concepts/ml/engineering/hf_inference_pydantic.md

"""
Exercise: Standardizing Hugging Face Inference with Pydantic
============================================================
Goal: Define typed I/O schemas for a text-classification pipeline,
then write a function that validates the raw pipeline output into
those schemas.

No real HF model is loaded — a mock pipeline is provided so you can
run this without a GPU or internet access.
"""

from pydantic import BaseModel, Field


# ---------------------------------------------------------------------------
# Mock pipeline (do not modify)
# ---------------------------------------------------------------------------


def _mock_pipeline(text: str) -> list[dict]:
    """Simulates the list-of-dicts a real HF pipeline returns."""
    return [{"label": "POSITIVE", "score": 0.9987}]


# ---------------------------------------------------------------------------
# TODO 1: Define InferenceRequest
# ---------------------------------------------------------------------------
# A Pydantic model with two fields:
#   - text: str  (the input sentence)
#   - model_id: str  (default: "distilbert-base-uncased-finetuned-sst-2-english")
#
class InferenceRequest(BaseModel):
    text: str
    model_id: str = "distilbert-base-uncased-finetuned-sst-2-english"


# ---------------------------------------------------------------------------
# TODO 2: Define SentimentResult
# ---------------------------------------------------------------------------
# A Pydantic model representing one item from the pipeline output:
#   - label: str
#   - score: float  — constrained to [0.0, 1.0]
# Hint: use Field(ge=..., le=...) for the constraint.
#
class SentimentResult(BaseModel):
    label: str
    score: float = Field(ge=0.0, le=1.0)


# ---------------------------------------------------------------------------
# TODO 3: Define InferenceResponse
# ---------------------------------------------------------------------------
# A Pydantic model that bundles:
#   - request: InferenceRequest  (echo the original request)
#   - results: list[SentimentResult]
#
class InferenceResponse(BaseModel):
    pass  # replace with your fields


# ---------------------------------------------------------------------------
# TODO 4: Implement run_inference
# ---------------------------------------------------------------------------
# Steps:
#   1. Call _mock_pipeline with request.text
#   2. Unpack each raw dict into a SentimentResult
#   3. Return an InferenceResponse wrapping the request and results
#
def run_inference(request: InferenceRequest) -> InferenceResponse:
    pass  # replace with your implementation


# ---------------------------------------------------------------------------
# Validation — do not modify
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    # Happy path
    req = InferenceRequest(text="The model deployment went smoothly!")
    resp = run_inference(req)
    print("Response:", resp)
    assert resp.results[0].label == "POSITIVE"
    assert 0.0 <= resp.results[0].score <= 1.0
    assert resp.request.text == req.text
    print("All assertions passed.")

    # Boundary check: invalid score should raise a ValidationError
    from pydantic import ValidationError

    try:
        SentimentResult(label="POSITIVE", score=1.5)  # score > 1.0
        print("ERROR: should have raised ValidationError")
    except ValidationError:
        print("ValidationError raised correctly for score=1.5.")

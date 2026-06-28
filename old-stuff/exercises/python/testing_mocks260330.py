import pytest
from unittest.mock import MagicMock

class HFInferenceClient:
    """Simulates an external call to HuggingFace API."""
    def predict(self, text: str) -> dict:
        # Imagine this makes a slow HTTP request
        raise NotImplementedError("Real API call should not be executed in tests!")

class ModelPipeline:
    def __init__(self, client: HFInferenceClient):
        self.client = client
        
    def classify_text(self, text: str) -> str:
        """
        Calls the client and parses the result.
        If the prediction score is > 0.8, return 'high_confidence'
        Otherwise, return 'low_confidence'
        """
        response = self.client.predict(text)
        if response.get("score", 0) > 0.8:
            return "high_confidence"
        return "low_confidence"

# --- Exercise ---
# 1. Create a pytest fixture named `mock_hf_client` that returns a MagicMock of HFInferenceClient.
@pytest.fixture
def mock_hf_client():
    return MagicMock(spec=HFInferenceClient)

# 2. Write `test_classify_text_high_confidence` using the fixture, mocking the predict method to return {"score": 0.9}.
def test_classify_text_high_confidence(mock_hf_client):
    pipeline = ModelPipeline(mock_hf_client)
    mock_hf_client.predict.return_value = {"score":0.9}
    result = pipeline.classify_text("hello world")
    assert result == "high_confidence"
    mock_hf_client.predict.assert_called_once_with("hello world")

# 3. Write `test_classify_text_low_confidence` using the fixture, mocking the predict method to return {"score": 0.5}.
def test_classify_text_low_confidence(mock_hf_client):
    pipeline = ModelPipeline(mock_hf_client)
    mock_hf_client.predict.return_value = {"score":0.1}
    result = pipeline.classify_text("helllo")
    assert result == "low_confidence"
    mock_hf_client.predict.assert_called_once_with("helllo")

# 4. Verify that the correct string is returned and that 'predict' was called with the correct argument.

# Keywords to search: "pytest fixture", "unittest.mock MagicMock return_value", "MagicMock assert_called_once_with"

# Your code below:

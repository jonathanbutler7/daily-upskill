# CONCEPT: Testing Service Interactions with Pytest and Mocks

> **Related Exercise**: `exercises/testing_mocks260330.py`

## When to Use

Testing code that depends on external services (databases, APIs, third-party libraries) without actually calling those services.

## Core Concept

**Pytest fixtures** provide reusable test setup, similar to `beforeEach` in Jest. **MagicMock** creates fake objects that record how they were used and can be configured to return specific values or raise exceptions.

## Pattern

```python
import pytest
from unittest.mock import MagicMock

@pytest.fixture
def mock_client():
    return MagicMock()

def test_service_call(mock_client):
    mock_client.get_data.return_value = {"id": 1, "name": "Test"}
    
    result = mock_client.get_data(1)
    
    assert result["name"] == "Test"
    mock_client.get_data.assert_called_once_with(1)
```

## Real-World DE/ML Scenario

Testing a data pipeline that fetches data from an external API. Instead of making real network calls (slow, flaky, costly), you mock the API client to return predictable data.

```python
def test_get_formatted_status(mock_client):
    mock_client.get_user_data.return_value = {"id": 1, "status": "Online"}
    
    service = UserService(mock_client)
    result = service.get_formatted_status(1)
    
    assert result == "User 1 is Online"
    mock_client.get_user_data.assert_called_once_with(1)
```

## Comparison with TypeScript (Optional)

| Python | TypeScript (Jest) |
|--------|-------------------|
| `@pytest.fixture` | `beforeEach` |
| `MagicMock()` | `jest.fn()` |
| `mock.return_value` | `mockFn.mockReturnValue()` |
| `assert_called_once_with()` | `expect(mockFn).toHaveBeenCalledWith()` |

## Keywords to Search

- "pytest fixtures vs setup_method"
- "unittest.mock MagicMock return_value"
- "mocking external API python"
- "pytest-mock vs unittest.mock"

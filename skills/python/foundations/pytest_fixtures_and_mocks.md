# Testing Service Interactions with pytest Fixtures and Mocks

When you write code that calls external services (APIs, databases, etc.), you need tests that verify your logic WITHOUT making actual network calls. This is where mocking comes in.

## Why Mock?

| Without Mock | With Mock |
|-------------|-----------|
| Tests call real APIs | Tests are fast and deterministic |
| Requires network + API keys | Runs offline, no setup |
| Can fail due to external issues | Fails only if YOUR code is broken |
| Rate limits, timeouts, downtime | You control the response |

## Pytest Fixtures

Fixtures provide test data or setup/teardown. They're more powerful than `setUp()`/`tearDown()` methods because they're reusable and can depend on other fixtures.

```python
import pytest

@pytest.fixture
def sample_config():
    """Returns test data - runs before each test that uses it."""
    return {"timeout": 30, "retries": 3}

def test_config(sample_config):
    assert sample_config["timeout"] == 30
```

Fixtures can also have setup/teardown with yield:

```python
@pytest.fixture
def database_connection():
    """Setup before test, teardown after."""
    conn = connect_to_test_db()
    yield conn  # Test runs here
    conn.close()  # Cleanup
```

## Mocking with unittest.mock.patch

`patch` replaces a function/class with a Mock object during the test:

```python
from unittest.mock import patch, Mock

@patch('requests.get')  # Patch where it's imported, not where it's defined
def test_api_call(mock_get):
    # Configure the mock
    mock_get.return_value = Mock(status_code=200, json=lambda: {"key": "value"})
    
    # Now requests.get() returns our fake response
    result = my_service.fetch_data()
    
    # Verify it was called correctly
    mock_get.assert_called_once_with("https://api.example.com/data")
```

## Key Concepts

### 1. Where to Patch
**Rule**: Patch where the object is *used*, not where it's *defined*.

```python
# my_service.py
import requests

def fetch():
    return requests.get("url")  # Patch 'my_service.requests.get'

# test_my_service.py
@patch('my_service.requests.get')  # ✓ Correct
def test_fetch(mock_get):
    pass

@patch('requests.get')  # ✗ Wrong - might not work as expected
def test_fetch(mock_get):
    pass
```

### 2. Configuring the Mock

```python
# Return a simple value
mock_get.return_value = "fake response"

# Return a mock with attributes
mock_get.return_value = Mock(status_code=200, json=lambda: {"data": "value"})

# Raise an exception
mock_get.side_effect = TimeoutError("Connection timed out")

# Call different values on successive calls
mock_get.side_effect = [Mock(status_code=200), Mock(status_code=500)]
```

### 3. Verifying Calls

```python
mock_get.assert_called_once_with("https://api.example.com")
mock_get.assert_called()  # Was called at least once
mock_get.call_count  # How many times called
mock_get.call_args  # Arguments of last call
```

## pytest-mock Plugin (Recommended)

The `pytest-mock` package provides a cleaner `mocker` fixture:

```bash
pip install pytest-mock
```

```python
def test_with_mocker(mocker):
    # Same as @patch, but cleaner syntax
    mock_get = mocker.patch('my_service.requests.get')
    mock_get.return_value = Mock(status_code=200)
    
    # Auto-cleanup - no need for context manager or finally block
    result = my_service.fetch_data()
```

**Benefits**:
- No context managers needed
- Automatic cleanup after test
- Idiomatic pytest style

## Real-World Pattern: Testing a Service

```python
# config_service.py
import requests

class ConfigService:
    def __init__(self, base_url: str):
        self.base_url = base_url
    
    def get_config(self, key: str) -> dict | None:
        try:
            response = requests.get(f"{self.base_url}/config/{key}")
            response.raise_for_status()
            return response.json()
        except (requests.HTTPError, requests.Timeout):
            return None
```

```python
# test_config_service.py
from unittest.mock import patch, Mock
import pytest

class TestConfigService:
    @patch('config_service.requests')
    def test_get_config_success(self, mock_requests):
        mock_response = Mock(status_code=200, json=lambda: {"timeout": 30})
        mock_requests.get.return_value = mock_response
        
        service = ConfigService("https://api.example.com")
        result = service.get_config("settings")
        
        assert result == {"timeout": 30}
        mock_requests.get.assert_called_once_with("https://api.example.com/config/settings")
    
    @patch('config_service.requests')
    def test_get_config_404_returns_none(self, mock_requests):
        mock_requests.get.return_value = Mock(status_code=404)
        mock_requests.HTTPError = requests.HTTPError  # Need to set this
        
        service = ConfigService("https://api.example.com")
        result = service.get_config("nonexistent")
        
        assert result is None
    
    @patch('config_service.requests')
    def test_get_config_timeout_returns_none(self, mock_requests):
        mock_requests.get.side_effect = TimeoutError()
        
        service = ConfigService("https://api.example.com")
        result = service.get_config("settings")
        
        assert result is None
```

## Comparison with Go

In Go, you'd typically use interfaces + mock implementations:

```go
// Interface defines what we need
type ConfigFetcher interface {
    GetConfig(key string) (*Config, error)
}

// Real implementation
type RealFetcher struct{}

func (f *RealFetcher) GetConfig(key string) (*Config, error) {
    // Makes real HTTP call
}

// Mock for testing
type MockFetcher struct{}

func (f *MockFetcher) GetConfig(key string) (*Config, error) {
    return &Config{Timeout: 30}, nil
}

// Test injects mock
func TestGetConfig(t *testing.T) {
    fetcher := &MockFetcher{}
    service := NewService(fetcher)
    // test service using fetcher
}
```

| Aspect | Python (Mock) | Go (Interface) |
|--------|--------------|----------------|
| Setup | `patch()` or `mocker.patch()` | Create mock struct |
| Cleanup | Auto with pytest-mock | Manual |
| Type checking | Via Protocol/typing | Via interface (static) |

## When to Use Each

- **Mock** external libraries (requests, boto3, redis)
- **Stub** your own classes for unit tests
- **Fake** for integration tests (in-memory DB instead of real DB)
- **Don't mock** your own domain logic - test that directly

## Installing Dependencies

```bash
pip install pytest pytest-mock requests
```

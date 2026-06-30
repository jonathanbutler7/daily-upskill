"""
Exercise: Testing Service Interactions with pytest Fixtures and Mocks

In real applications, your code often talks to external services (APIs, databases, etc.).
In unit tests, you want to test YOUR logic without actually calling those services.
This is where mocking comes in.

Real-world scenario: 
You have a ConfigService that fetches settings from an external API.
You need to test that your service handles the response correctly,
without making actual HTTP calls during testing.

Keywords to research:
- pytest fixtures
- unittest.mock.patch
- pytest.mark.parametrize
- Mock return values
"""

import pytest
from unittest.mock import Mock, patch


# --- Part 1: The Service We're Testing ---

class ConfigService:
    """
    A service that fetches configuration from an external API.
    
    In production, this would make real HTTP calls.
    In tests, we'll mock the HTTP client.
    """
    
    def __init__(self, api_base_url: str):
        self.api_base_url = api_base_url
    
    def get_config(self, config_key: str) -> dict | None:
        """
        Fetch a config value from the external API.
        
        TODO: Implement this method to:
        1. Make an HTTP GET request to {api_base_url}/config/{config_key}
        2. Return the JSON response as a dict
        3. Return None if the request fails (404, timeout, etc.)
        
        HINT: You'll need to use `requests.get()` or `httpx.get()`
        """
        # Your implementation here
        pass
    
    def get_all_configs(self) -> list[dict]:
        """
        Fetch all configurations.
        
        TODO: Implement this method to:
        1. Make an HTTP GET request to {api_base_url}/config
        2. Return a list of config dicts
        3. Return empty list if the request fails
        """
        # Your implementation here
        pass


# --- Part 2: The Tests ---

class TestConfigService:
    """Tests for ConfigService with mocked HTTP calls."""
    
    @pytest.fixture
    def mock_api_response(self):
        """Fixture that provides a mock for the HTTP library."""
        # TODO: Use patch to create a mock for requests.get (or httpx.get)
        # HINT: with patch('requests.get') as mock_get:
        pass
    
    def test_get_config_returns_dict_on_success(self, mock_api_response):
        """
        Test that get_config returns a dict when the API returns 200.
        
        TODO: Set up the mock to return a successful response,
        then assert that get_config returns the expected dict.
        """
        # Setup: Configure the mock
        # mock_api_response.return_value.json.return_value = {"key": "value"}
        # mock_api_response.return_value.status_code = 200
        
        # Execute
        # service = ConfigService("https://api.example.com")
        # result = service.get_config("feature_flags")
        
        # Assert
        # assert result == {"key": "value"}
        pass
    
    def test_get_config_returns_none_on_404(self, mock_api_response):
        """
        Test that get_config returns None when the API returns 404.
        
        TODO: Set up the mock to raise an HTTPError or return 404,
        then assert that get_config returns None.
        """
        pass
    
    def test_get_config_returns_none_on_timeout(self, mock_api_response):
        """
        Test that get_config handles timeouts gracefully.
        
        TODO: Use mock to simulate a timeout exception,
        then assert that get_config returns None.
        """
        pass


# --- Part 3: Stretch Goal ---

class TestConfigServiceWithPytestMock:
    """
    Alternative approach using pytest-mock plugin.
    
    This requires: pip install pytest-mock
    
    The `mocker` fixture provides a cleaner syntax for patching.
    """
    
    def test_get_config_with_mocker_fixture(self, mocker):
        """
        Same test as before, but using pytest-mock.
        
        TODO: Implement the same test using mocker.patch() instead of unittest.mock.patch()
        
        Benefits of pytest-mock:
        - Cleaner: mocker.patch() vs with patch() context manager
        - Auto-cleanup: mocks are automatically cleaned up after the test
        - Consistent with pytest style
        """
        pass


# --- Run Instructions ---
# To run these tests:
# 1. Install pytest: pip install pytest pytest-mock requests
# 2. Run: python -m pytest exercises/pytest_fixtures_mocks_260328.py -v

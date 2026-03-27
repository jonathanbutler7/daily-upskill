import functools
import random
import time

# --- Exercise: Retry Decorator for a Flaky API ---
# Task: Complete the @retry decorator below to automatically retry a function 
# that may fail. It should catch exceptions and retry up to `max_retries` 
# before raising the error.

def retry(max_retries=3):
    """
    A decorator that retries a function a specified number of times if it fails.
    
    Keywords to search for:
    - "Python decorator with arguments"
    - "functools.wraps"
    - "Python try-except in a loop"
    """
    def decorator(func):
        # Hint: Use @functools.wraps(func) on your wrapper function
        # to preserve the original function's metadata!
        
        def wrapper(*args, **kwargs):
            # 1. Start a loop for max_retries
            # 2. In a try block, call the original function (func(*args, **kwargs))
            # 3. If it succeeds, return the result
            # 4. If it fails (except), log the failure and retry
            # 5. If it reaches the final retry and still fails, raise the exception
            i = 1
            while i <= max_retries:
                try:
                    result = func(*args, **kwargs)
                except:
                    if i > max_retries:
                        print("Failed!")
                        return
                    else:
                        i += 1
            return result
        return wrapper
    
    return decorator

# --- Flaky API Simulation (Don't Change This Part) ---

@retry(max_retries=3)
def fetch_user_data(user_id: int):
    """Simulates a call to an external user service that fails 50% of the time."""
    print(f"DEBUG: Attempting to fetch data for user {user_id}...")
    
    if random.choice([True, False]):
        raise ConnectionError(f"Failed to reach user service for ID {user_id}")
    
    return {"id": user_id, "name": "Alice"}

if __name__ == "__main__":
    try:
        data = fetch_user_data(123)
        print(f"SUCCESS: Fetched data: {data}")
    except ConnectionError as e:
        print(f"CRITICAL: Failed to fetch data after multiple attempts: {e}")

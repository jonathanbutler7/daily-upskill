"""
Async/Await Exercise
====================
Goal: Understand how async/await enables concurrent I/O operations
and compare performance with synchronous code.

Instructions:
1. Implement the TODO sections below
2. Run the script and observe the timing differences
3. Answer the reflection questions at the bottom

Reference: concepts/python/concurrency/async_await.md
"""

import asyncio
import time
from typing import Any


# =============================================================================
# TASK 1: Implement an async function that simulates an API call
# =============================================================================
async def fetch_user_data(user_id: int, latency: float = 0.5) -> dict:
    """
    Simulate fetching user data from an API.
    
    TODO:
    1. Print a message indicating the fetch has started (include user_id)
    2. Use `await asyncio.sleep(latency)` to simulate network delay
    3. Print a message indicating the fetch completed
    4. Return a dict with user_id and some mock data
    
    Args:
        user_id: The ID of the user to fetch
        latency: Simulated network latency in seconds
        
    Returns:
        Dict containing user data, e.g., {"user_id": 1, "name": "User 1", "status": "active"}
    """
    # YOUR CODE HERE
    print(f"Fetch started for user: {user_id}")
    await asyncio.sleep(latency)
    print(f"Fetch completed for user: {user_id}")
    return {"user_id": user_id, "name": "Tim", "status": "active"}


# =============================================================================
# TASK 2: Implement a sync version for comparison
# =============================================================================
def fetch_user_data_sync(user_id: int, latency: float = 0.5) -> dict:
    """
    Synchronous version of fetch_user_data for comparison.
    
    TODO: Same as above, but use time.sleep() instead of asyncio.sleep()
    
    Args:
        user_id: The ID of the user to fetch
        latency: Simulated network latency in seconds
        
    Returns:
        Dict containing user data
    """
    # YOUR CODE HERE
    print(f"Fetch started for user: {user_id}")
    time.sleep(latency)
    print(f"Fetch completed for user: {user_id}")
    return {"user_id": user_id, "name": "Tim", "status": "active"}


# =============================================================================
# TASK 3: Fetch multiple users sequentially (async but not concurrent)
# =============================================================================
async def fetch_users_sequential(user_ids: list[int]) -> list[dict]:
    """
    Fetch multiple users one at a time (sequential awaits).
    
    TODO:
    1. Create an empty results list
    2. Loop through user_ids
    3. Await fetch_user_data for each user (one at a time)
    4. Append each result to the list
    5. Return the results
    
    Args:
        user_ids: List of user IDs to fetch
        
    Returns:
        List of user data dicts
    """
    # YOUR CODE HERE
    results = []
    for user_id in user_ids:
        user = await fetch_user_data(user_id)
        results.append(user)
    
    return results
        
        


# =============================================================================
# TASK 4: Fetch multiple users concurrently
# =============================================================================
async def fetch_users_concurrent(user_ids: list[int]) -> list[dict]:
    """
    Fetch multiple users concurrently using asyncio.gather().
    
    TODO:
    1. Create a list of coroutines (one fetch_user_data call per user_id)
    2. Use asyncio.gather() to run them all concurrently
    3. Return the results
    
    Hint: asyncio.gather(*coroutines) unpacks the list
    
    Args:
        user_ids: List of user IDs to fetch
        
    Returns:
        List of user data dicts
    """
    # YOUR CODE HERE
    coroutines = []
    for user_id in user_ids:
        coroutines.append(fetch_user_data(user_id))
    result = await asyncio.gather(*coroutines)    
    return result
        
        
    


# =============================================================================
# TASK 5: Implement a timing wrapper for async functions
# =============================================================================
async def time_async_execution(coro) -> tuple[Any, float]:
    """
    Execute a coroutine and return both its result and execution time.
    
    TODO:
    1. Record start time
    2. Await the coroutine
    3. Record end time
    4. Return tuple of (result, elapsed_time)
    
    Args:
        coro: A coroutine to execute
        
    Returns:
        Tuple of (coroutine_result, elapsed_time_in_seconds)
    """
    # YOUR CODE HERE
    start = time.time()
    result = await(coro)
    end = time.time()
    return (result, end - start)


# =============================================================================
# MAIN: Run the benchmarks
# =============================================================================
async def main():
    print("=" * 60)
    print("Async/Await Performance Comparison")
    print("=" * 60)
    
    # Configuration
    USER_IDS = [1, 2, 3, 4, 5]
    LATENCY = 0.5  # seconds per "API call"
    
    print(f"\nFetching {len(USER_IDS)} users, each with {LATENCY}s latency")
    print(f"Expected sequential time: {len(USER_IDS) * LATENCY}s")
    print(f"Expected concurrent time: ~{LATENCY}s")
    
    # ----- SYNCHRONOUS BENCHMARK -----
    print("\n--- SYNCHRONOUS (baseline) ---")
    # TODO: Time fetching all users synchronously using fetch_user_data_sync
    start = time.time()
    for user_id in USER_IDS:
        data = fetch_user_data_sync(user_id, LATENCY)
    end = time.time()
    elapsed = end - start
    print(f"Sync time: {elapsed:.2f}s")
    
    # ----- ASYNC SEQUENTIAL BENCHMARK -----
    print("\n--- ASYNC SEQUENTIAL ---")
    # TODO: Time fetch_users_sequential
    start = time.time()
    end = time.time()
    elapsed = end - start
    # fetch_users_sequential()
    print(f"Async sequential time: {elapsed:.2f}s")
    
    # ----- ASYNC CONCURRENT BENCHMARK -----
    print("\n--- ASYNC CONCURRENT ---")
    # TODO: Time fetch_users_concurrent
    start = time.time()
    end = time.time()
    elapsed = end - start
    fetch_users_concurrent()
    # print(f"Async concurrent time: {elapsed:.2f}s")
    
    print("\n" + "=" * 60)
    print("REFLECTION QUESTIONS (answer in comments below)")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())


# =============================================================================
# REFLECTION QUESTIONS
# =============================================================================
"""
After running the benchmarks, answer these questions:

1. Why was async sequential about the same speed as synchronous?
   
   YOUR ANSWER:
   

2. Why was async concurrent so much faster?
   
   YOUR ANSWER:
   

3. What does asyncio.gather() do under the hood?
   
   YOUR ANSWER:
   

4. When would you use async/await instead of threading for I/O-bound work?
   
   YOUR ANSWER:
   

5. Can you use async/await for CPU-bound work? Why or why not?
   
   YOUR ANSWER:
   
"""

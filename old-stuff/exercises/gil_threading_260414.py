"""
GIL & Threading Exercise
=========================
Goal: Understand when the GIL helps vs hurts performance by comparing
different concurrency approaches for CPU-bound and I/O-bound tasks.

Instructions:
1. Implement the TODO sections below
2. Run the script and observe the timing differences
3. Answer the reflection questions at the bottom

Reference: concepts/python/concurrency/gil_and_threading.md
"""

import threading
import multiprocessing
import time
from typing import Callable, Any


# =============================================================================
# TASK 1: Implement a CPU-bound function
# =============================================================================
def cpu_intensive_work(iterations: int) -> int:
    """
    Perform CPU-intensive computation.
    
    TODO: Implement a loop that does some math work (e.g., sum of squares).
    This should keep the CPU busy without any I/O.
    
    Args:
        iterations: Number of iterations to perform
        
    Returns:
        The computed result (e.g., sum of squares)
    """
    # YOUR CODE HERE
    pass


# =============================================================================
# TASK 2: Implement an I/O-bound function
# =============================================================================
def io_bound_work(delay_seconds: float) -> str:
    """
    Simulate I/O-bound work (like an API call or file read).
    
    TODO: Use time.sleep() to simulate waiting for I/O.
    
    Args:
        delay_seconds: How long to "wait" for I/O
        
    Returns:
        A string indicating completion
    """
    # YOUR CODE HERE
    pass


# =============================================================================
# TASK 3: Implement timing utilities
# =============================================================================
def time_execution(func: Callable, *args, **kwargs) -> tuple[Any, float]:
    """
    Execute a function and return both its result and execution time.
    
    TODO: Record start time, call the function, record end time, return both.
    
    Args:
        func: The function to execute
        *args, **kwargs: Arguments to pass to the function
        
    Returns:
        Tuple of (function_result, elapsed_time_in_seconds)
    """
    # YOUR CODE HERE
    pass


# =============================================================================
# TASK 4: Run tasks with threading
# =============================================================================
def run_with_threads(target: Callable, args_list: list[tuple], num_threads: int) -> float:
    """
    Run a function multiple times using threads.
    
    TODO:
    1. Create `num_threads` Thread objects, each calling `target` with args from `args_list`
    2. Start all threads
    3. Wait for all threads to complete (join)
    4. Return total elapsed time
    
    Args:
        target: The function each thread will run
        args_list: List of argument tuples, one per thread
        num_threads: Number of threads to create
        
    Returns:
        Total elapsed time in seconds
    """
    # YOUR CODE HERE
    pass


# =============================================================================
# TASK 5: Run tasks with multiprocessing
# =============================================================================
def run_with_processes(target: Callable, args_list: list[tuple], num_processes: int) -> float:
    """
    Run a function multiple times using separate processes.
    
    TODO:
    1. Create `num_processes` Process objects, each calling `target` with args from `args_list`
    2. Start all processes
    3. Wait for all processes to complete (join)
    4. Return total elapsed time
    
    Hint: multiprocessing.Process works similarly to threading.Thread
    
    Args:
        target: The function each process will run
        args_list: List of argument tuples, one per process
        num_processes: Number of processes to create
        
    Returns:
        Total elapsed time in seconds
    """
    # YOUR CODE HERE
    pass


# =============================================================================
# MAIN: Run the benchmarks
# =============================================================================
def main():
    print("=" * 60)
    print("GIL & Threading Performance Comparison")
    print("=" * 60)
    
    # Configuration
    CPU_ITERATIONS = 5_000_000  # Adjust if too slow/fast on your machine
    IO_DELAY = 1.0  # seconds
    NUM_WORKERS = 4
    
    # ----- CPU-BOUND BENCHMARKS -----
    print("\n--- CPU-BOUND TASKS ---")
    print(f"Running {NUM_WORKERS} tasks, each with {CPU_ITERATIONS:,} iterations\n")
    
    # Single-threaded (baseline)
    # TODO: Time running cpu_intensive_work NUM_WORKERS times sequentially
    # print(f"Single-threaded:  {elapsed:.2f}s")
    
    # Multi-threaded
    # TODO: Time running cpu_intensive_work with threads
    # print(f"Multi-threaded:   {elapsed:.2f}s")
    
    # Multi-process
    # TODO: Time running cpu_intensive_work with processes
    # print(f"Multi-process:    {elapsed:.2f}s")
    
    # ----- I/O-BOUND BENCHMARKS -----
    print("\n--- I/O-BOUND TASKS ---")
    print(f"Running {NUM_WORKERS} tasks, each with {IO_DELAY}s delay\n")
    
    # Single-threaded (baseline)
    # TODO: Time running io_bound_work NUM_WORKERS times sequentially
    # print(f"Single-threaded:  {elapsed:.2f}s")
    
    # Multi-threaded
    # TODO: Time running io_bound_work with threads
    # print(f"Multi-threaded:   {elapsed:.2f}s")
    
    print("\n" + "=" * 60)
    print("REFLECTION QUESTIONS (answer in comments below)")
    print("=" * 60)


if __name__ == "__main__":
    main()


# =============================================================================
# REFLECTION QUESTIONS
# =============================================================================
"""
After running the benchmarks, answer these questions:

1. For CPU-bound tasks, was multi-threading faster than single-threaded?
   Why or why not?
   
   YOUR ANSWER:
   

2. For CPU-bound tasks, was multi-processing faster than single-threaded?
   Why or why not?
   
   YOUR ANSWER:
   

3. For I/O-bound tasks, was multi-threading faster than single-threaded?
   Why or why not?
   
   YOUR ANSWER:
   

4. When would you choose threading over multiprocessing in a real project?
   
   YOUR ANSWER:
   

5. What's one downside of using multiprocessing compared to threading?
   
   YOUR ANSWER:
   
"""

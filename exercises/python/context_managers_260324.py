"""
Task: Efficient Data Handling with Context Managers & File I/O
Date: 2026-03-24

Scenario:
You have a "large" dataset (simulated) containing sensor readings.
Your task is to:
1. Use a context manager to safely open and read the file.
2. Process the file line-by-line (to simulate memory efficiency).
3. Calculate the average value of a specific sensor metric.
4. Write the result to a new file, also using a context manager.

Keywords to search for:
- "Python context manager with statement"
- "Python read file line by line memory efficiency"
- "Python csv module DictReader"
"""

import csv
import os

def generate_dummy_data(filename: str, num_rows: int = 1000):
    """Generates a dummy CSV file if it doesn't exist."""
    if not os.path.exists(filename):
        with open(filename, 'w', newline='') as f:
            writer = csv.writer(f)
            writer.writerow(['timestamp', 'sensor_id', 'value'])
            for i in range(num_rows):
                writer.writerow([f"2026-03-24T12:{i//60:02}:{i%60:02}", "temp_01", 20.0 + (i % 10)])
        print(f"Created {filename}")

def calculate_average_sensor_value(input_file: str, sensor_id: str) -> float:
    """
    Reads the input_file and calculates the average 'value' for the given 'sensor_id'.
    TODO: Implement using a context manager and line-by-line reading.
    """
    total = 0.0
    count = 0
    
    # --- YOUR CODE HERE ---
    # Hint: Use 'with open(...) as f:'
    # Hint: Consider using csv.DictReader(f)
    with open(input_file, 'r') as file:
        reader =  csv.DictReader(file)
        for row in reader:
            if row['sensor_id'] == sensor_id:
                count += 1
                total = total + float(row['value'])
    return total / count
    # ----------------------

def save_result(output_file: str, result: float):
    """
    Writes the result to the output_file.
    TODO: Implement using a context manager.
    """
    # --- YOUR CODE HERE ---
    with open(output_file, "w") as file:
        file.write(str(result))
    # ----------------------

if __name__ == "__main__":
    DATA_FILE = "sensor_readings.csv"
    RESULT_FILE = "average_temp.txt"
    
    generate_dummy_data(DATA_FILE)
    
    avg_temp = calculate_average_sensor_value(DATA_FILE, "temp_01")
    print(f"Calculated average temperature: {avg_temp}")
    
    if avg_temp is not None:
        save_result(RESULT_FILE, avg_temp)
        print(f"Result saved to {RESULT_FILE}")

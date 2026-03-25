from functools import reduce

# Scenario: You're cleaning up experiment results before feeding them into a report.
# You need to filter out failed experiments, extract the accuracies, and calculate a total.

results = [
    {"experiment": "A1", "accuracy": 0.85, "status": "COMPLETED"},
    {"experiment": "B2", "accuracy": 0.42, "status": "COMPLETED"},
    {"experiment": "C3", "accuracy": 0.91, "status": "FAILED"},
    {"experiment": "D4", "accuracy": 0.76, "status": "COMPLETED"},
    {"experiment": "E5", "accuracy": 0.33, "status": "COMPLETED"},
]

def run_pipeline(data):
    # 1. Filter: Keep only experiments where status is "COMPLETED"
    completed = filter(lambda x: x['status'] == "COMPLETED", data) # TODO: Replace None with your lambda

    # 2. Map: Extract just the 'accuracy' float from each experiment dict
    accuracies = map(lambda x: x['accuracy'], completed) # TODO: Replace None with your lambda
    # print(list(accuracies))
    # 3. Reduce: Sum all the accuracies together
    # Hint: reduce(lambda acc, x: ..., accuracies, 0)
    total_accuracy = reduce(lambda acc, x: acc + x, accuracies, 0) # TODO: Replace None with your lambda

    return total_accuracy

if __name__ == "__main__":
    # Once implemented, this should print 2.36 (0.85 + 0.42 + 0.76 + 0.33)
    total = run_pipeline(results)
    print(f"Total Accuracy of Completed Experiments: {total:.2f}")

    # Exercise 2 (Mental or Code):
    # How would you use reduce to find the MAX accuracy instead of the sum?

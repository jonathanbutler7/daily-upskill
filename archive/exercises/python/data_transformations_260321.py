"""
Daily Exercise: Practical Data Transformations
Goal: Use list comprehensions to filter and format a list of transaction records.
Scenario: You're cleaning a list of transaction data from a financial service.

Task:
1. Filter the `transactions` list to include only those where the 'status' is 'successful'.
2. Format each successful transaction as a string: "Transaction ID: {id} - Amount: ${amount}"
3. Use a list comprehension to accomplish this in a single line if possible.

Hint: Use string formatting (f-strings) inside the comprehension's expression.
Keywords to search for: "python list comprehension", "python f-strings"
"""

transactions = [
    {"id": 101, "amount": 55.0, "status": "successful"},
    {"id": 102, "amount": 12.5, "status": "failed"},
    {"id": 103, "amount": 120.0, "status": "successful"},
    {"id": 104, "amount": 9.99, "status": "successful"},
    {"id": 105, "amount": 3.0, "status": "pending"},
]

# TODO: Create the 'formatted_transactions' list using a list comprehension
formatted_transactions = [f"Transaction ID: {x["id"]} - Amount: ${x["amount"]}" for x in transactions if x["status"] == "successful"]

# Print the results to verify
if __name__ == "__main__":
    print("Formatted Successful Transactions:")
    for tx in formatted_transactions:
        print(tx)

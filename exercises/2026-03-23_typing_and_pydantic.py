"""
Daily Exercise: Robust Data Validation with Pydantic
Goal: Define a schema and validate incoming data.
Scenario: You're ingesting transaction data from an external API that sometimes sends malformed records.

Task:
1. Define a Pydantic model named `Transaction` with:
   - `id`: int
   - `amount`: float (must be > 0)
   - `status`: Literal['successful', 'failed', 'pending']
2. Loop through the `raw_data` and try to validate each record with the model.
3. Catch `ValidationError` for malformed records and print a helpful message.
4. Store the valid objects in `validated_transactions`.

Hint: Use `from pydantic import BaseModel, Field, ValidationError`.
Keywords to search for: "pydantic basemodel", "pydantic validation error", "pydantic Field constraints"
"""

from pydantic import BaseModel, Field, ValidationError
from typing import Literal

# 1. TODO: Define your Transaction model here
class Transaction(BaseModel):
    id: int
    amount: float = Field(..., gt=0) # Must be greater than zero
    status: Literal['successful', 'failed', 'pending']

raw_data = [
    {"id": 201, "amount": 100.50, "status": "successful"},
    {"id": "202", "amount": 15.0, "status": "failed"},  # Should be valid (coercion)
    {"id": 203, "amount": -5.0, "status": "successful"}, # Should fail (amount <= 0)
    {"id": 204, "amount": 22.0, "status": "pending"},
    {"id": 205, "amount": 9.99, "status": "error"},     # Should fail (invalid status)
]

validated_transactions = []

# 2. TODO: Loop through raw_data and validate each record
# Note: Use Transaction(**record) to create the model instance

for transaction in raw_data:
    try:
        validated_transaction = Transaction(**transaction)
        validated_transactions.append(validated_transaction)
    except ValidationError as e:
        print("Got an error")


if __name__ == "__main__":
    print(f"Successfully validated {len(validated_transactions)} transactions.")
    for tx in validated_transactions:
        print(tx)

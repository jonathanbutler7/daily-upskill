# Ledger DB Scenarios

Use this file to reason through the ledger behavior before writing the schema or Go code.

For each scenario, fill in:

- Starting balances
- Attempted business action
- Ledger transaction
- Ledger entries
- Invariants
- Should it post or fail?
- Why?
- Ending balances
- What could go wrong with concurrency?

## Scenario 1: Alice Sends Bob $10 And Has $20

Starting balances:

```text
Alice: $20
Bob:   $0
```

Attempted business action:

```text
Alice sends Bob $10.
```

Ledger transaction:

```text
type: transfer
from: Alice
to: Bob
amount: $10
status: posted
```

Ledger entries:

```text
Alice wallet   -10
Bob wallet     +10
------------------
Total            0
```

Invariants:

```text
1. Sum of entries for the posted transaction must equal 0.
2. Alice's balance must not go below 0.
3. The transaction must have at least two entries.
```

Should it post or fail?

```text
Post.
```

Why?

```text
The entries balance:
-10 + 10 = 0

Alice has enough money:
20 - 10 = 10
```

Ending balances:

```text
Alice: $10
Bob:   $10
```

What could go wrong with concurrency?

```text
If another request spends Alice's money at the same time, both requests might read Alice's starting balance as $20.

That is okay only if the final result is still valid. If the combined spending would make Alice negative, one transaction should fail or retry.
```

## Scenario 2: Alice Sends Bob $10 And Has $5

Starting balances:

```text
Alice: $5
Bob: $0
```

Attempted business action:

```text
Alice sends Bob $10.
```

Ledger transaction:

```text
none
```

Ledger entries:

```text
none
```

Invariants:

```text
1. Sum of entries for posted transaction must equal 0.
2. Transferring a higher amount than the user has results in error
3. Transaction must only have one entry, Bob's account is not reached
```

Should it post or fail?

```text
fail
```

Why?

```text
Alice has insufficient funds
```

Ending balances:

```text
Alice: $5
Bob: $0
```

What could go wrong with concurrency?

```text
If Alice receives money at the same time as trying to send it, the error could be incorrect because Alice actually had the money she needed when transferring the money to Bob
```

## Scenario 3: Alice Sends Bob $10 Twice With The Same Idempotency Key

Starting balances:

```text
Alice: $10
Bob: $0
```

Attempted business action:

```text
Alice sends Bob $10 twice. First request posts, but the second request returns the first result. Does not create a second transaction.
```

Ledger transaction:

```text
type: transfer
from: Alice
to: Bob
amount: $10
status: posted
```

Ledger entries:

```text
Alice wallet   -10
Bob wallet     +10
------------------
Total            0
```

Invariants:

```text
1. Sum of entries for posted transaction must equal 0.
2. Transactions must be idempotent so multiple attempts on the same request do not result in double transactions
3. Transaction must have 2 entries
```

Should it post or fail?

```text
post, but only once
```

Why?

```text
Alice has enough money to send 1 transfer to bob for $10.

The entries balance: -10 + 10 = 0

Alice has enough money: 10 - 10 = 0

The system is idempotent and will prevent multile transactions from getting Alice's wallet into a bad state. 

The following will not happen to Alice: 

10 - 10 - 10 = -10
```

Ending balances:

```text
Alice: $0
Bob: $10
```

What could go wrong with concurrency?

```text
In this case, concurrency is handled by an idempotency key on the request, so the same request run twice does not result in two seaparte requests being run
```

## Scenario 4: Alice Has $10; Two Concurrent Requests Each Send $8

Starting balances:

```text
Alice: $10
```

Attempted business action:

```text
Alice sends two concurrent requests for $8. 

This can happen by
- Double clicking send button
- Mobile retry after timeout
```

Ledger transaction:

```text
type: transfer
from: Alice
to: Unspecified
amount: $8
status: posted

type: transfer
from: Alice
to: Unspecified
amount: $8
status: failed_insufficient_funds
```

Ledger entries:

```text
Alice wallet              -8
Someone else's wallet     +8
----------------------------
Total                      0
```

Invariants:

```text
1. 2 concurrent requests must be handled separately
2. If 1 request is good, only allow one. If two are good, allow both, but fail when the request amount exceeds the wallet balance
3. No balances below zero
```

Should it post or fail?

```text
one should post, one should fail
```

Why?

```text
$10 - $8 = $2
$2 - $8 = $-6 (no balances below zero)
```

Ending balances:

```text
Alice: $2
Someone else: $8
```

What could go wrong with concurrency?

```text
It could go wrong if the actions are both handled and Alice gets to a balance below zero.
```

## Scenario 5: Alice Sends Bob $10, Then The Transfer Is Reversed

Starting balances:

```text
Alice: $10
Bob:   $0
```

Attempted business action:

```text
Alice sends Bob $10. Then the support team reverses the transaction and sends $10 back to Alice
```

Ledger transaction:

```text
type: transfer
from: Alice
to: Bob
amount: $10
status: posted

type: reversal
from: Bob
to: Alice
amount: $10
status: posted
```

Ledger entries:

```text
TRANSFER
Alice wallet   -10
Bob wallet     +10
------------------
Total            0

REVERSAL
Bob wallet     -10
Alice wallet   +10
------------------
Total            0
```

Invariants:

```text
1. Each posted transaction must balance to $0.
2. Reversal must relate to origianl transaction.
3. Original transaction not edited or deleted
4. Transaction can only be reversed once
5. Reversal should not make Bob's balance negative
```

Should it post or fail?

```text
Both the transaction and the reversal should post
```

Why?

```text
The entries balance:
-10 + 10 + -10 + 10 = 0

Alice has enough money:
10 - 10 = 0

Bob has enough for the reversal:
10 - 10 = 0
```

Ending balances:

```text
Alice: $10
Bob:   $0
```

What could go wrong with concurrency?

```text
If bob spends $10 the same time support tries to reverse the transfer
```

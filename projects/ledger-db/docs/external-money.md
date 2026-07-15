# External Money

External money includes moving money to or from a payment rail outside the ledger db system.

## Direction

Direction will be one of 
- `deposit`
- `withdrawal`

### `deposit`

Deposits occur when a user moves money from an external account into the ledger db system.

### `withdrawal`

Withdrawals occur when a user moves money from the ledger db system to an external account.

## Maintaining the ledger

In the case of moving money to or from an external account, the ledger db system doesn't have full knowledge of the external account details. Therefore ledger db keeps an "Cash settlement" account to keep double entry records and track the payment rails event for every external transfer.

### `external_reference`

Identifier provided by the payment rail that allows us to have an reference to how they identify the transfer within their own system.

## Status

Since this is currently a POC, deposits and withdrawals are posted as `posted` immediately. 

Future work can post as `pending` and then receive updates from the payment rails when the transfer settles, and handle retries and reconciliation.

## Reconciliation

TODO: Reconciliation checks whether the rail event and ledger activity agree.

Examples:
- an external transfer exists but has no linked ledger transaction
- a deposit or withdrawal transaction exists without an external transfer row
- the external transfer status does not match the ledger state
- the same external reference is reused
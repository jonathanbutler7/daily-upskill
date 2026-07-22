# Weave Data Model Mapping

This maps the Weave payments tables from the pasted schema onto the current `ledger-db` model.

The short version:

- Weave merchants map to ledger accounts.
- Weave patients are external payers, not ledger accounts.
- Weave invoices are payment requests or receivables, not ledger transactions by themselves.
- Weave payment logs are payment attempts/results.
- A settled successful payment maps to a ledger deposit.
- `invoice_payments` links the invoice request to the payment event that funded it.

## Current Ledger Model

The current ledger model has three core accounting tables:

- `ledger_accounts`: account balances inside the ledger.
- `ledger_transactions`: business events that move money.
- `ledger_entries`: accounting lines for each transaction.

External payment information lives outside the accounting entries:

- `external_transfers`: payment rail metadata such as external reference, rail, status, direction, and linked ledger transaction.

That separation matters. The ledger entries say how money moved inside the ledger. The external transfer row says what outside payment event caused it.

## Weave Tables

### `payments.invoices`

An invoice is a request for a patient to pay a merchant.

Important fields:

- `id`: invoice id.
- `merchantid`: dentist office / merchant receiving the payment.
- `personid`: patient identity.
- `amount`: requested payment amount.
- `status`: invoice lifecycle state.
- `uniquelink`: patient-facing payment link.
- patient fields: payer context, not ledger account data.
- surcharge fields: pricing/payment amount details.

Ledger mapping:

- Does not directly become a ledger transaction.
- Provides payment intent/context.
- `merchantid` determines which ledger account receives funds.
- `id` is useful as invoice metadata and reconciliation context.

### `payments.payment_log`

A payment log row is closer to the actual payment event.

Important fields:

- `id`: payment event id.
- `merchantid`: dentist office / merchant receiving the payment.
- `amount`: amount paid.
- `weavefee`: platform fee, if applicable.
- `paymentstatus`: payment lifecycle state.
- `confirmationcode`: processor-facing reference or confirmation.
- `submittedat`, `recordedat`, `createdat`: lifecycle timestamps.
- `processorid`: processor reference.
- `paymenttype`: card, ACH, or another payment method category.
- `processortype`: processor name, currently defaulting to `stripe`.

Ledger mapping:

- Successful settled payment becomes a `deposit` ledger transaction.
- `payment_log.id` should be the idempotency key for replay.
- `confirmationcode` or `processorid` should become `external_transfers.external_reference`.
- `paymenttype` / `processortype` should map to `external_transfers.rail`.
- `paymentstatus` should map to `external_transfers.status`.

### `payments.invoice_payments`

This table links invoice records to payment log records.

Important fields:

- `invoiceid`: invoice request.
- `paymentid`: payment event.
- `invoice_change_id`: invoice change/version context.

Ledger mapping:

- This is reconciliation glue.
- It answers: which payment funded which invoice?
- It should not become a ledger entry.
- It can be used to prove every posted invoice payment has a matching ledger deposit.

## Main Happy Path

When a patient pays a dentist office invoice and the payment is settled:

```text
Weave:
payments.invoices
  merchantid = dentist office
  personid = patient
  amount = 10000

payments.payment_log
  id = payment event id
  merchantid = same dentist office
  amount = 10000
  paymentstatus = successful/settled
  processorid or confirmationcode = processor reference

payments.invoice_payments
  invoiceid = invoice id
  paymentid = payment_log id
```

Maps to:

```text
ledger_transactions
  type = deposit
  idempotency_key = payment_log.id
  from_account_id = Cash Settlement
  to_account_id = dentist office ledger account
  amount = payment_log.amount
  currency_code = USD

ledger_entries
  Cash Settlement: -10000
  Dentist Office: +10000

external_transfers
  direction = deposit
  rail = mapped paymenttype/processortype
  status = posted
  external_reference = processorid or confirmationcode
  user_account_id = dentist office ledger account
  ledger_transaction_id = deposit transaction id
  amount = payment_log.amount
  currency_code = USD
```

The patient does not need a ledger account for this flow. The patient is the external payer. The dentist office is the account holder inside the ledger.

## Concept Mapping

| Weave concept | Current ledger concept | Notes |
| --- | --- | --- |
| `payments.invoices.merchantid` | `ledger_accounts.id` through a merchant-to-account mapping | The current ledger needs a lookup table or import map because ledger account ids are local integers. |
| Dentist office | Ledger account holder | This is the balance owner. |
| Patient / `personid` | External payer metadata | Do not model as a normal ledger account for settled invoice payments. |
| Invoice | Payment request / receivable context | Useful for reconciliation, but not itself money movement. |
| Payment log | External payment event | This is the best source for creating deposits. |
| `invoice_payments` | Reconciliation link | Connects the request to the payment event. |
| `payment_log.id` | `ledger_transactions.idempotency_key` | Stable replay key. Prevents duplicate posting. |
| `processorid` / `confirmationcode` | `external_transfers.external_reference` | Stable processor reference, if present. |
| `paymentstatus` | `external_transfers.status` | Needs an explicit status mapping. |
| `paymenttype` / `processortype` | `external_transfers.rail` | Current ledger only allows `ach` and `instant`, so this needs widening for card/stripe/etc. |
| `amount` | `ledger_transactions.amount` and `external_transfers.amount` | Must be minor units. Confirm whether Weave stores cents. |
| `weavefee` | Future fee transaction or fee entries | Current ledger does not model fees yet. |
| Surcharge fields | Future surcharge metadata or separate fee/surcharge entries | Do not fold this into the base payment amount without deciding gross vs net accounting. |

## What Fits Without Big Changes

The core payment flow fits well:

```text
patient pays invoice -> external payment event -> dentist office ledger deposit
```

The current ledger already has the important pieces:

- Merchant/dentist office can be represented by `ledger_accounts`.
- Successful payment can be represented by a `deposit` transaction.
- The processor-side event can be represented by `external_transfers`.
- Duplicate replay can use `payment_log.id` as the idempotency key.
- The entries stay balanced through the `Cash Settlement` system account.

## Gaps To Handle Before Replay

### Merchant Account Mapping

Weave uses UUID merchant ids. The ledger uses `bigserial` account ids.

For replay, create an import map like:

```text
weave_merchant_id -> ledger_account_id
```

This can be a temporary CSV, a staging table, or a future `merchant_accounts` table.

### Payment Status Mapping

Do not post every `payment_log` row.

Create an explicit mapping from Weave `paymentstatus` values to ledger behavior:

```text
successful/settled -> post deposit
pending/authorized -> skip for now, or record external transfer only
failed/expired/canceled -> do not post ledger entries
refunded/charged back -> reversal, withdrawal, or adjustment flow
```

The exact integer meanings need to come from Weave code/docs before using real data.

### Rails

The current ledger allows:

```text
ach
instant
```

Weave has:

```text
paymenttype
processortype
```

For realistic replay, `external_transfers.rail` probably needs values like:

```text
card
ach
stripe
manual
```

Or split processor from rail later:

```text
rail = card | ach
processor = stripe | ...
```

### Fees And Surcharges

The current ledger can post the gross payment into the dentist office account, but it does not yet model fees cleanly.

Possible accounting choices:

1. Gross deposit first, fee later:

```text
Cash Settlement: -10000
Dentist Office: +10000

Dentist Office: -300
Weave Fee Revenue / Processor Fees: +300
```

2. Net deposit only:

```text
Cash Settlement: -9700
Dentist Office: +9700
```

For learning, gross deposit plus explicit fee transaction is better because it explains what happened.

### Refunds And Chargebacks

Refunds and chargebacks should not edit the original payment.

They should become one of:

- a reversal of the original deposit, if it fully cancels the payment
- a withdrawal/refund transaction, if money is sent back out
- an adjustment transaction, if only part of the payment is corrected

The right choice depends on how Weave records refund and chargeback states.

## Replay Contract

A safe replay input should be smaller than the full production schema.

Start with sanitized rows shaped like:

```text
payment_id
invoice_id
merchant_id
amount
currency
payment_status
payment_type
processor_type
processor_reference
submitted_at
recorded_at
```

Do not include patient names, emails, phone numbers, birthdates, chart numbers, memo text, or raw processor payloads.

## First Replay Rule

For the first version, only replay settled successful payments.

Pseudo-flow:

```text
for each payment_log row:
  if payment status is not settled/successful:
    record as skipped
    continue

  find ledger account for merchantid

  call PostExternalTransfer(deposit):
    UserAccountID = dentist office ledger account
    TransferAmount = payment_log.amount
    Rail = mapped payment type
    ExternalReference = processorid or confirmationcode or payment_log.id
    IdempotencyKey = payment_log.id
    ExternalTransferDirection = deposit
```

Then verify:

- every posted payment has one ledger transaction
- every posted ledger deposit has one external transfer
- stored balances match derived balances
- replaying the same file a second time does not double-post
- failed/pending/canceled payments do not create ledger entries
- payments for unknown merchants fail clearly

## Summary

The Weave model fits the current ledger model if the dentist office is the ledger account holder and the patient is external payment context.

The biggest changes are not fundamental ledger changes. They are import/reconciliation additions:

- merchant UUID to ledger account mapping
- Weave payment status mapping
- richer rail/processor values
- fee/surcharge modeling
- refund/chargeback handling

The first useful stress test should replay only settled successful payments into dentist office ledger accounts.

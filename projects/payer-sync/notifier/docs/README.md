# Notifier

## Overview

The notifier module handles the final part of the payer sync pipeline by reading the status of a payment whose writeback has reached its terminal state, whether that is `POSTED`, `PARTIALLY_POSTED` OR `WRITEBACK_FAILED`, and sending an appropriate notification based on the status. `POSTED` and `PARTIALLY_POSTED` result in an email to the office while `WRITEBACK_FAILED` results in an email to the operations team to address the issue.

## Invariants
- The notifier system is only concerned with writebacks that have reached their terminal state.
- Notification failure must never change the payment writeback state
- `POSTED` is the happy path and results in the notifier sending an email to the office with payment details.
- `PARTIALLY_POSTED` partial success state indicating that some of the PMS ledger methods were successful but some failed. Notifier sends an email to the office with payment details.
- `WRITEBACK_FAILED` failure state indicating the payment was processed, but writing back to the PMS failed, even after retrying 3 times. 

## Audience

- Operations: internal staff who, in the event of a writeback failure, may be able to intervene and help resolve a writeback failure
- Office: office staff who would be only interested in the business outcome of payer sync. Broader audience than operations. 

## Idempotency

Notifier must use stable values to generate an idempotency key to prevent duplicate notifications.

Suggested format:
- office email: `payersync:notify:{office_id}:{reconciled_payment_id}:{status}`
- ops alert: `payersync:notify:ops:{reconciled_payment_id}:{status}`

## Failure

- SMTP failure must result in retry independent of writebacks
- Notification retries must not result in duplicate sending

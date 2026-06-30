# Data Readiness Checklist

Purpose: train on what we have, not what we wish we had.

Source schemas reviewed:
- db/payments/payments.md
- db/persons/persons.md
- db/schedule/schedule.md

Status legend:
- PRESENT: fields exist and are usable now.
- PARTIAL: fields exist but need mapping/rules/cleanup.
- MISSING: not in current schemas.

Criticality legend:
- 🔴: required to train/evaluate a valid paid-within-30-days model.
- 🟠: strong impact on model quality and business usefulness.
- 🟡: useful lift or segmentation, but model can run without it.
- ⚠️: nice-to-have context with limited direct impact.

## 1) Core Label Requirements (Paid Within 30 Days)

- [x] Criticality: 🔴 | invoice principal and created timestamp
  - payments.invoices.amount
  - payments.invoices.createdat
- [ ] PARTIAL | Criticality: 🟡 | invoice due/expiration semantics
  - payments.invoices.expiresat exists, but we still need exact semantics (payment-link expiry vs due date)
- [x] Criticality: 🔴 | payment success timestamp
  - payments.payment_log.submittedat / recordedat exist and are used in current label generation.
- [x] Criticality: 🔴 | direct invoice-to-payment linkage
  - join table is `payments.invoice_payments`
- [x] Criticality: 🔴 | status code dictionaries
- [-] Criticality: 🟠 | reversal semantics
  - refund/chargeback/void rules and fields are not identified in provided schemas
  - leave refunds/chargebacks out for now; this can be a future feature

## 2) Unified Feature Readiness (Including docs/data-points-to-consider.md)

### Financial and insurance context (ability to pay)

- [ ] Criticality: 🟠 | insurance coverage type (self-pay/commercial/Medicaid)
  - no insurance plan/type fields found in payments/persons.
- [x] Criticality: 🟠 | total out-of-pocket balance
  - payments.invoices.amount exists and is usable as a proxy.
- [ ] Criticality: 🟠 | patient responsibility ratio (patient owed / total charge)
  - no insurance-paid amount or total charge breakdown identified.

### Historical and behavioral data (willingness to pay)

- [x] Criticality: 🟠 | patient tenure
  - public.person.entry_date (with created_at fallback).
- [ ] PARTIAL | Criticality: 🟡 | appointment reliability (no-show rate)
  - scheduler.calendars_events.attendee_status includes NO_SHOW,
  - still needs mapping rules to person/invoice timeline for model-safe features.
- [x] Criticality: 🟠 | historical payment speed (median days-to-pay)
  - derived from invoice created timestamp and first successful payment timestamp.

### Service and clinical context (urgency)

- [ ] PARTIAL | Criticality: 🟡 | procedure type
  - scheduler.appointment_types.display_name/description provides an appointment-type proxy,
  - no clinical procedure/encounter coding fields found.
- [x] Criticality: 🟡 | future appointments scheduled
  - scheduler.calendars_events.start_date/start_time and attendee_id support future-appointment flags.

### Schedule context available now

- [x] Criticality: 🟡 | appointment lifecycle signals
  - scheduler.calendars_events.attendee_status, type, recurring, recurrence_rule.
- [x] Criticality: 🟡 | booking intent and intake context
  - scheduler.booking_submissions.requested_slots, booking_source, reviewed_status, created_at.

### Temporal and targeting features

- [x] Criticality: 🟡 | payday temporal feature
  - derive day-of-month from payments.invoices.createdat.
- [x] Criticality: 🟠 | cold-start feature (is_new_patient)
  - derive from public.person.entry_date/created_at relative to invoice timestamp.
- [x] Criticality: 🔴 | strict target definition (days_to_payment <= 30)
  - clock start exists (payments.invoices.createdat),
  - but invoice-payment linkage can be found in payments.invoice_payments

### Invoice/payment context available now

- [x] Criticality: 🟠 | invoice timestamps and expiry
  - payments.invoices.createdat, expiresat.
- [x] Criticality: 🟡 | payment channel/process hints
  - payments.payment_log.paymenttype, origin, processortype.
- [x] Criticality: ⚠️ | surcharge context
  - payments.invoices.appliedsurchargepercentage, appliedsurchargeamount, desiredsurchargepercentage, surchargingenabled.

### Person/contact/location context available now

- [x] Criticality: 🟠 | ZIP and location proxies for Census enrichment
  - public.address.postal_code,
  - public.client_location.postal_code, city, state.
- [x] Criticality: 🟡 | contactability signals
  - public.contact_info.type, destination, normalized_destination, priority,
  - payments.invoices.personemail, personmobilephone, personhomephone (fallback).
- [x] Criticality: 🟡 | household/account structure
  - public.person.household_id,
  - public.account_relationship.person_id, account_id, relationship.

### Additional high-value features still missing

- [ ] PARTIAL | Criticality: 🟠 | reminder/outreach timeline (dunning events)
  - scheduler.calendars_events.type includes REMINDER,
  - no billing-collection/dunning-specific event log identified.
- [ ] Criticality: 🟠 | financing/payment-plan offer + acceptance

## 4) Data Quality Rules To Enforce In Synthetic Generation

- [ ] TODO: respect not-null and defaults from schema (especially ids, timestamps, booleans)
- [ ] TODO: keep enum domains valid once codebooks are provided
- [ ] TODO: enforce temporal order
  - invoice.createdat <= payment.submittedat <= payment.recordedat (when present)
- [ ] TODO: generate realistic null patterns
  - optional contact/person fields should not be fully populated for every row
- [ ] TODO: preserve key relationships
  - invoices.personid must map to person.id when non-null
  - address.person_id one-to-one assumption should match unique index behavior

## 5) Immediate Next Data Requests

Priority 0 (blockers for reliable labels):
1. Definition of "successful payment"

Priority 1 (high-value features):
1. Insurance/patient responsibility fields
2. Link scheduling events to invoice/person timeline for feature engineering
3. Collection workflow/reminder events

## 6) Go / No-Go For Model Proof

- Go for early signal demo if:
  - we can link invoices to payments,
  - we have status code mappings,
  - we freeze a clear paid_within_30_days label contract.

Current status:
- invoice-to-payment linkage: satisfied (`payments.invoice_payments`)
- status mappings: satisfied for current baseline (`paymentstatus == 1` as successful)
- paid_within_30_days label contract: implemented and validated in baseline notebook

- No-go for claims about production readiness if any of the above are missing.

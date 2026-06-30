# Collection Recommendation v1

Purpose: add a deterministic next-best-action layer on top of payment prediction so office staff can collect more revenue with consistent workflows.

## 1. Recommendation payload contract

The model score stays the same. The recommendation is a policy output generated from score + account context.

```json
{
  "patient_id": "p_123",
  "location_id": "loc_1",
  "prediction": {
    "will_pay_in_30": false,
    "confidence_score": 0.22,
    "risk_band": "low",
    "scored_at": "2026-06-10T17:30:00Z",
    "model_version": "will_they_pay_v1"
  },
  "recommendation": {
    "policy_version": "collection_policy_v1",
    "priority_rank": 8,
    "expected_value_at_risk": 620.00,
    "action_code": "CALL_AND_OFFER_PLAN",
    "channel": "phone",
    "due_by": "2026-06-11T17:00:00Z",
    "fallback_action_code": "SEND_SMS_PAY_LINK",
    "fallback_after_days": 2,
    "reason_codes": [
      "LOW_PAYMENT_PROBABILITY",
      "HIGH_OPEN_BALANCE",
      "NO_RECENT_PAYMENT"
    ],
    "script_hint": "Patient has high outstanding balance and low pay likelihood. Offer a 3-part plan and collect first payment today."
  }
}
```

## 2. Fields useful to office staff

Show these fields in UI first:

1. risk band and confidence
2. expected dollars at risk
3. one primary action for today
4. fallback action if no response
5. top reason codes (max 3)
6. due-by time for follow-up

Do not show raw feature values or model internals.

## 3. Deterministic policy rules (v1)

Inputs required:

- `confidence_score` (0.0 to 1.0)
- `risk_band` (`low`, `medium`, `high` likelihood to pay)
- `open_balance_amount`
- `days_since_last_payment`
- `has_payment_method_on_file`

Derived values:

- `non_payment_probability = 1 - confidence_score`
- `expected_value_at_risk = open_balance_amount * non_payment_probability`

Rules:

1. If `risk_band = low` and `open_balance_amount >= 400`:
- action: `CALL_AND_OFFER_PLAN`
- channel: `phone`
- fallback: `SEND_SMS_PAY_LINK` after 2 days

2. If `risk_band = low` and `open_balance_amount < 400`:
- action: `SEND_SMS_PAY_LINK`
- channel: `sms`
- fallback: `CALL_REMINDER` after 3 days

3. If `risk_band = medium` and no payment method on file:
- action: `REQUEST_CARD_ON_FILE`
- channel: `phone`
- fallback: `SEND_EMAIL_PLAN_OPTIONS` after 3 days

4. If `risk_band = medium` and payment method exists:
- action: `STANDARD_REMINDER_SEQUENCE`
- channel: `sms`
- fallback: `CALL_REMINDER` after 4 days

5. If `risk_band = high`:
- action: `LIGHT_TOUCH_REMINDER`
- channel: `sms`
- fallback: `NONE`

6. Escalation override: if `expected_value_at_risk >= 1000`, set action to `MANAGER_REVIEW` regardless of other rules.

## 4. Priority queue ordering

Daily outreach queue sort key:

1. `expected_value_at_risk` descending
2. then `due_by` ascending
3. then `confidence_score` ascending

This orders work by recoverable revenue, not just model probability.

## 5. Reason-code generation (simple and auditable)

Recommend max 3 codes:

- `LOW_PAYMENT_PROBABILITY` when `confidence_score < 0.40`
- `HIGH_OPEN_BALANCE` when `open_balance_amount >= 400`
- `NO_RECENT_PAYMENT` when `days_since_last_payment >= 60`
- `NO_CARD_ON_FILE` when `has_payment_method_on_file = false`

## 6. v1 rollout and measurement

Primary success metric:

- incremental cash collected vs control workflow

Operational metrics:

- days-to-payment
- promise-to-pay conversion rate
- collection touches per dollar collected
- payment-plan acceptance rate

Guardrails:

- no increase in complaint rate
- no increase in bad-debt write-off rate

## 7. Minimal implementation plan

1. Add policy function in batch scorer or API read path.
2. Store recommendation JSON alongside prediction output.
3. Expose recommendation fields in prediction response.
4. Render one action card in collections UI.
5. Log recommendation served, action taken, and payment outcome.

The policy is deterministic so offices can audit and trust recommendations before moving to a learned recommendation system.

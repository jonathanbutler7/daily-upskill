# Will They Pay? \- Technical Product Document

## 1\. Overview

Goal: show payment-likelihood predictions in the Weave payments UI.

Architecture:

1. Batch scoring pipeline (Python): load patient data, run trained model, write predictions.  
2. Schema service endpoint (Go gRPC/REST): read stored prediction for a patient.

The model does not run at request time. The Go service reads from a pre-scored predictions store.

**Current implementation:** predictions are written to `data/output/scored_predictions_prod.csv` and the Go service reads from that file, configured via `WTP_PREDICTIONS_CSV_PATH`. This is a placeholder for the `payment_predictions` Postgres table described in section 4. The CSV approach eliminates the database dependency for the initial rollout; swap in the Postgres write path when the table is provisioned.

---

## 2\. Inference strategy

### Chosen approach: pre-scored batch \+ stored predictions

Batch runs, scores patients with open balances, writes predictions, and the schema service reads from the predictions store.

> **Current state:** batch is run manually and writes to a CSV file (`data/output/scored_predictions_prod.csv`). The diagram below reflects the target production state.

> **Future state:** automated nightly batch job writing to `payment_predictions` Postgres table.

```
[Python batch job]
  reads patient + payment data from Postgres
  joins census income data by zipcode
  runs trained model
  writes to payment_predictions
         |
         v
[payment_predictions]
         |
         v
[Go schema service]
  SELECT ... WHERE patient_id = $1
         |
         v
[Payments UI]
```

### Why not real-time scoring

Real-time scoring (Vertex AI or Python service per request) adds:

- always-on model-serving infrastructure  
- API latency impact  
- more provisioning/monitoring complexity

Too risky for hackathon scope.

Production direction: Vertex AI fallback on cache miss (see section 3).

### Freshness tradeoff

Batch scoring can be stale. Patients created after the last run may be unscored.

Hackathon behavior: return no prediction for unscored patients; UI shows empty state.

Production options:

- Option A: hourly batch (\~60 minute max staleness), no new runtime infra.  
- Option B: hybrid: batch for known patients \+ synchronous Vertex call on cache miss.

---

## 3\. Schema service contract

### New RPC: `GetPaymentLikelihood`

Request:

```protobuf
message GetPaymentLikelihoodRequest {
  string patient_id  = 1;
  string location_id = 2;
}
```

Response:

```protobuf
message GetPaymentLikelihoodResponse {
  PaymentLikelihood likelihood = 1;
  CollectionRecommendation recommendation = 2;
}

message PaymentLikelihood {
  bool   has_prediction     = 1;  // false if no score exists
  bool   will_pay_in_30     = 2;
  float  confidence_score   = 3;  // 0.0 - 1.0
  string risk_band          = 4;  // "high" | "medium" | "low"
  string model_version      = 5;
  string scored_at          = 6;  // ISO 8601 timestamp
}

message CollectionRecommendation {
  bool   has_recommendation       = 1;
  string policy_version           = 2;
  int32  priority_rank            = 3;
  float  expected_value_at_risk   = 4;  // open_balance * (1 - confidence)
  string action_code              = 5;  // CALL_AND_OFFER_PLAN, SEND_SMS_PAY_LINK, ...
  string channel                  = 6;  // phone | sms | email
  string due_by                   = 7;  // ISO 8601 timestamp
  string fallback_action_code     = 8;
  int32  fallback_after_days      = 9;
  repeated string reason_codes    = 10; // max 3
  string script_hint              = 11; // staff-facing call guidance
}
```

Behavior:

- If `has_prediction = false`, frontend shows "no prediction available" and no score.  
- `risk_band` is derived from `confidence_score`:  
  - `>= 0.70` \-\> `"high"` (likely to pay)  
  - `0.40 - 0.69` \-\> `"medium"`  
  - `< 0.40` \-\> `"low"` (unlikely to pay)  
- `model_version` tracks which model artifact generated the score (debugging and future A/B).
- `recommendation` is generated from deterministic policy rules in
  `model/docs/collection-recommendation-v1.md`.
- If `has_prediction = true` and `has_recommendation = false`, UI shows score only and no action card.

Errors:

- `NOT_FOUND`: patient\_id does not exist.  
- `INTERNAL`: database read failure.

---

## 4\. Predictions table schema

```sql
CREATE TABLE payment_predictions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          TEXT NOT NULL,
    location_id         TEXT NOT NULL,
    will_pay_in_30      BOOLEAN NOT NULL,
    confidence_score    NUMERIC(5, 4) NOT NULL,  -- e.g. 0.8231
    risk_band           TEXT NOT NULL,
    model_version       TEXT NOT NULL,
    feature_snapshot    JSONB,                   -- optional model inputs used
    scored_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ON payment_predictions (patient_id, location_id, model_version);
CREATE INDEX ON payment_predictions (patient_id);
CREATE INDEX ON payment_predictions (scored_at);
```

Notes:

- `feature_snapshot` keeps prediction input features for debugging.  
- Unique key `(patient_id, location_id, model_version)` enables upsert on reruns for the same model version.  
- Go service reads latest row per patient: `ORDER BY scored_at DESC LIMIT 1`.

---

## 5\. Batch scoring pipeline

Implementation: Python script (Colab for hackathon, local or scheduled job for production).

Steps:

```
1. Connect to Postgres.
2. Pull patients with open balances.
   SELECT patient_id, location_id, zipcode, balance_amount,
          historical_payment_count, historical_on_time_rate,
          last_payment_age_days
   FROM   <patient + balance views>
   WHERE  balance_amount > 0

3. Join census data by zipcode (CSV or table).
   Adds: census_median_income, census_household_size

4. Build feature matrix.
5. Load trained model artifact (.pkl or .joblib).
6. Run model.predict() and model.predict_proba().
7. Derive risk_band from confidence score.
8. Upsert into payment_predictions:
   INSERT INTO payment_predictions (...) VALUES (...)
   ON CONFLICT (patient_id, location_id, model_version)
   DO UPDATE SET
     will_pay_in_30   = EXCLUDED.will_pay_in_30,
     confidence_score = EXCLUDED.confidence_score,
     risk_band        = EXCLUDED.risk_band,
     feature_snapshot = EXCLUDED.feature_snapshot,
     scored_at        = now()
```

Model artifact:

```py
import joblib
joblib.dump(model, "will_they_pay_v1.joblib")
```

Hackathon: local file. Production: GCS or model registry.

Feature inputs at scoring time:

| feature | source |
| :---- | :---- |
| `balance_amount` | Postgres |
| `historical_payment_count` | Postgres |
| `historical_on_time_rate` | Postgres |
| `last_payment_age_days` | Postgres |
| `census_median_income` | zipcode CSV or table |
| `census_household_size` | zipcode CSV or table |

Training and scoring must match exactly on feature names, order, and null handling. Share one feature-builder function between notebook and batch script.

---

## 6\. Training pipeline (Colab)

Training remains in Colab.

```
1. Load historical patient + payment data.
2. Join census data.
3. Build labels: paid_within_30_days.
4. Build feature matrix (same as section 5).
5. Train and evaluate candidates (logistic regression, random forest, XGBoost).
6. Select best model by ROC-AUC.
7. Save model artifact: joblib.dump(model, "will_they_pay_v1.joblib").
8. Save feature column list: feature_columns.json.
```

`feature_columns.json` is required. It defines exact model input order and must be loaded by the batch scorer.

---

## 7\. Data access and PII constraints

Batch job may read:

- opaque internal `patient_id` (not SSN, MRN, or name)  
- `location_id`  
- 5-digit zipcode  
- derived financial aggregates (balance amount, payment count, on-time rate, recency)  
- census aggregates by zipcode

Must not enter training or prediction storage:

- name, DOB, SSN, phone, email, address  
- raw appointment/service dates tied to patient ID  
- medical record numbers or health plan IDs

Use derived fields (for example `days_since_last_payment`) instead of raw timestamps (`last_payment_date`).

---

## 8\. UI behavior

Frontend behavior for `GetPaymentLikelihoodResponse`:

When `has_prediction = true`:

```
Payment likelihood
----------------------------------
Unlikely to pay within 30 days
Confidence: 78% | Last scored: 2h ago

Consider offering a payment plan or financing option
```

When `has_prediction = false`:

```
Payment likelihood
----------------------------------
No prediction available yet
```

Rules:

- Never show a score when `has_prediction = false`.  
- Show relative `scored_at` for freshness.  
- Do not expose `model_version` or raw feature inputs to office staff.  
- Action hint by `risk_band`:  
  - `low` \-\> "Consider offering a payment plan or financing option"  
  - `medium` \-\> "Standard collection flow recommended"  
  - `high` \-\> no hint

---

## 9\. Open questions

1. Which Postgres instance should batch use: prod read replica or separate analytics DB?  
2. Is there a shared GCS bucket for production model artifacts?  
3. Should this be a new top-level schema service method or live under an existing payments namespace?  
4. For hourly production batch, what scheduler should be used: k8s CronJob, Cloud Scheduler, or other?  
5. Should prediction history be retained per model version, or is latest-only overwrite acceptable?

---

## 10\. Recommendation layer (v1)

The office workflow should include a next-best-action recommendation in addition to
payment likelihood. This is a deterministic policy based on score + account context,
not a second ML model.

Goals:

- improve collections throughput by ranking work by expected dollars at risk
- provide a clear action for staff today and a fallback if no response
- keep recommendations auditable and easy to tune

Policy and field definitions are specified in:

- `model/docs/collection-recommendation-v1.md`

Initial rollout:

1. Generate recommendation in the same batch job that writes predictions.
2. Store recommendation payload (JSONB) with prediction rows.
3. Return recommendation from `GetPaymentLikelihood`.
4. Show one recommendation card in UI with action, reason codes, and due-by.


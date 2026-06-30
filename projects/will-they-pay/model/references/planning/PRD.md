# Tab 1

# Will They Pay? PRD

## 1\. Summary

Offices lack a payment-likelihood signal, so collection decisions are blind.

Build an ML risk score that predicts payment within 30 days and surface it in payments workflows.

Hackathon deliverables:

1. Extract patient, balance, and payment history data  
2. Enrich with zipcode-level census income data  
3. Train and compare baseline models  
4. Batch-score open balances  
5. Show prediction \+ confidence in UI  
6. Preserve path to 30/60/90-day prediction and financing recommendations

---

## 2\. Problem

At charge time, offices cannot estimate payment likelihood and use one workflow for all patients.

Impact:

- late or missed payments are not proactively handled  
- financing is not offered when needed  
- collections performance drops

A simple point-of-collection risk signal can improve decisions without manual analysis.

---

## 3\. Goals

### Business

- reduce revenue loss from late/non-payments  
- enable proactive financing offers for high-risk patients  
- prove Weave payments data has predictive value beyond transaction processing

### Product

- predict payment within 30 days  
- return a confidence score  
- surface results at point of collection  
- retain metadata for debugging and explainability

### Technical

- train on historical data when available  
- compare 2-3 baseline models  
- run feature joins offline (no real-time cross-DB queries)  
- scope for 2-day delivery

### Out of scope for now

- request-time scoring  
- online training/retraining  
- automated financing recommendations  
- production monitoring, drift detection, fairness audits  
- 30/60/90-day prediction (future phase)

---

## 4\. Users

**Office staff at point of collection** need a simple signal: "likely to pay" or "unlikely to pay" with confidence.

**Hackathon demo audience** Need to see realistic workflow integration and a credible path to production.

---

## 5\. Assumptions

1. Historical data is sufficient to label paid-within-30-days outcomes.  
2. Zipcode coverage is high enough to use as a feature.  
3. Census zipcode data can be joined offline.  
4. Payment history features are useful but optional for MVP if joins are slow/complex.  
5. Batch scoring is sufficient for demo; real-time inference is not required.  
6. UI can be lightweight or mocked if workflow is demonstrable.

---

## 6\. What the system does

For each patient balance/payment request, the system:

1. collect patient and financial attributes  
2. enrich with zipcode-level census data  
3. run trained model  
4. Outputs:  
   - `will_pay_in_30_days` — yes/no  
   - `confidence_score` — 0.0 to 1.0  
   - `risk_band` — high / medium / low (optional)  
5. store predictions for UI reads  
6. show result at collection decision time

### UI surface

Show a prediction card in:

- the collections modal  
- the payment request list  
- the payment request details page

Example copy:

```
Unlikely to pay within 30 days
Confidence: 78%
Consider offering a payment plan or financing option
```

---

## 7\. Pipeline overview

Pipeline stages:

1. data ingest: export patient, balance, payment data  
2. feature enrichment: join census zipcode data, derive model features  
3. label generation: mark paid-within-30-days outcomes  
4. model training: compare baselines, select best performer  
5. batch scoring: score open balances  
6. prediction storage: write table/file for UI reads  
7. UI display: show label, confidence, action hint

---

## 8\. Architecture

```
Operational data export     census.gov (zipcode income data)
         |                            |
         v                            v
    Raw dataset  ─────────────>  Feature builder
                                       |
                                       v
                               Labeled training set
                                       |
                                       v
                           Model training + evaluation
                           (logistic regression, random forest, XGBoost)
                                       |
                                       v
                               Selected model artifact
                                       |
                                       v
                             Batch scoring (Colab or script)
                                       |
                                       v
                               Predictions table / file
                                       |
                                       v
                             Payments UI  (collections modal, PR list)
```

### Approach

Hackathon approach is fully offline/batch:

- one data snapshot/export  
- train in Colab  
- save model artifact  
- score demo dataset  
- show stored predictions in UI

This avoids risky real-time cross-DB integration in a 2-day window.

### Stack

- **Python / Colab**: feature engineering, training, evaluation, scoring  
- **Go / existing API layer**: optional, serve stored predictions to UI  
- **Postgres or CSV**: intermediate datasets and final predictions

---

## 9\. Data model

### Training row

One row per patient-payment opportunity:

| field | description |
| :---- | :---- |
| `patient_id` |  |
| `location_id` |  |
| `zipcode` |  |
| `balance_amount` |  |
| `appointment_date` |  |
| `payment_request_created_at` |  |
| `historical_payment_count` | optional |
| `historical_on_time_rate` | optional |
| `last_payment_age_days` | optional |
| `census_median_income` | from zipcode join |
| `census_household_size` | from zipcode join |
| `label_paid_within_30_days` | true/false — the target |

### Prediction record

One row per scored item:

| field | description |
| :---- | :---- |
| `prediction_id` |  |
| `patient_id` |  |
| `payment_request_id` |  |
| `model_version` |  |
| `predicted_at` |  |
| `will_pay_in_30_days` |  |
| `confidence_score` |  |
| `risk_band` |  |
| `top_features_json` | optional, for debugging |

### Label definition

`paid_within_30_days = true` if the first successful payment timestamp is within 30 calendar days of payment request creation.

If payment request creation is unavailable, use the closest operational timestamp and document the substitution.

---

## 10\. Model approach

Binary classification. Start with simple baselines.

**Candidate models:**

1. logistic regression: fast, interpretable baseline  
2. random forest: handles nonlinear interactions  
3. XGBoost: often strongest if data quality/volume supports it

Select by ROC-AUC. Prioritize precision on the "unlikely to pay" class: false positives are costlier than false negatives.

**Features to try:**

- balance amount  
- payment history (count, on-time rate, recency)  
- zipcode-level income / household size  
- appointment timing (day of week, proximity to month end)  
- location-level payment rate aggregate

**Minimum evaluation output:**

- accuracy  
- precision / recall  
- ROC-AUC  
- confusion matrix

---

## 11\. Inference

MVP uses batch scoring only (one-off or periodic runs on open balances).

Move to request-time scoring later when:

- feature availability is stable and fast  
- there's actual latency budget defined  
- model serving infrastructure exists

---

## 12\. UI behavior

Do not show a raw score without context. Show label \+ confidence \+ simple action hint.

If no prediction exists, show nothing.

---

## 13\. Operational constraints

- scoring must be reproducible from the same snapshot  
- every prediction must include model version  
- inputs must be inspectable for debugging  
- safe fallback: show no prediction instead of fabricated output

---

## 14\. Data and compliance

- treat patient/payment data as sensitive  
- minimize PII in exports  
- zipcode demographic enrichment is acceptable when joined/stored correctly  
- prefer de-identified demo subsets when production access is uncertain

---

## 15\. Risks and open questions

1. Do we have enough labeled history to train something meaningful?  
2. What is the correct 30-day clock start: appointment, payment request creation, or other?  
3. Are payment history joins feasible at hackathon speed, or do we ship without them?  
4. Is zipcode income data strong enough as a feature on its own, or does it need payment history to be useful?  
5. Can we do a real table read from the UI, or do we need mocked predictions for the demo?

---

## 16\. MVP scope

**In:**

- binary prediction: will pay within 30 days  
- confidence score  
- batch scoring only  
- offline feature joins  
- 2-3 baseline models compared  
- prediction \+ confidence in UI with simple action hint

**Fallback if payment history joins are too slow:**

- train on balance amount \+ zipcode features only  
- add history fields in a later iteration

---

## 17\. Definition of done

Demo must show:

1. real (or realistic) historical dataset used for training/evaluation  
2. model comparison and model selection rationale  
3. scored open-balance set with prediction \+ confidence  
4. UI view of a specific patient in collection workflow  
5. extension story to 30/60/90-day prediction and financing recommendations

---

## 18\. What this is

Batch-scored payment risk system for healthcare offices: ingest data, enrich with census income, train classifier, score open balances, and surface results in collection workflows.

Core bet: a simple model using balance \+ zipcode income can improve collection decisions and evolve into richer recommendation workflows.

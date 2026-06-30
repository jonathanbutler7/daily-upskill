# Model Runtime Contract

Purpose: make it unambiguous which model drives predictions, what inputs are required, and what outputs are produced.

## 1) Which Model Is Used

- Training winner: XGBoost (`best_model = XGBoost`) in `models/eval_summary_prod.json`.
- Model version written into scored output: `will_they_pay_prod_v1`.
- Runtime serving behavior in Go service:
  - Service reads `model/data/output/scored_predictions_prod.csv` (or path from `WTP_PREDICTIONS_CSV_PATH`).
  - Service does not load `.joblib` directly at request time.
  - First row for a given `personid` is used when duplicates exist.

## 2) Inputs Required To Generate Predictions (Scoring Pipeline)

The production scoring notebook uses this feature set from `models/feature_columns_prod.json`:

1. amount
2. created_day_of_week
3. created_day_of_month
4. created_hour
5. payment_origin_mode
6. surchargingenabled
7. is_guardian
8. account_age_days
9. tenure_months
10. is_new_patient
11. average_days_to_pay
12. appointment_reliability_score
13. median_household_income
14. poverty_rate_pct
15. average_household_size
16. bachelors_or_higher_pct
17. unemployment_rate_pct

These are transformed in `notebooks/production/train_model_prod.ipynb` and then scored by the selected model.

## 3) Output Produced By Scoring Pipeline

Scoring output file: `data/output/scored_predictions_prod.csv`

Expected columns:
1. personid
2. postal_code
3. will_pay_in_30
4. confidence_score
5. risk_band
6. model_version
7. scored_at

Column notes:
- `will_pay_in_30`: binary prediction (0/1).
- `confidence_score`: probability of positive class.
- `risk_band`: low/medium/high derived from confidence thresholds.
- `model_version`: currently `will_they_pay_prod_v1`.
- `scored_at`: RFC3339 timestamp.

## 4) Inputs Required By Serving API

For `GetPaymentLikelihood` requests, service requires:
1. patient_id
2. location_id

Matching logic:
- `patient_id` must match `personid` in scored CSV.
- If no match exists, API returns an error.

## 5) Output Returned By Serving API

From `GetPaymentLikelihoodResponse`:
1. likelihood.has_prediction
2. likelihood.will_pay_in_30
3. likelihood.confidence_score
4. likelihood.risk_band
5. likelihood.model_version
6. likelihood.scored_at
7. recommendation fields (policy output)

## 6) Operational Rule

Whenever model training changes:
1. regenerate `models/feature_columns_prod.json` and `models/eval_summary_prod.json`.
2. regenerate `data/output/scored_predictions_prod.csv`.
3. ensure `model_version` in output matches the active model release.

# Definition of Done

Working checklist for PRD section 17.

Source of truth:
- [PRD Section 17](docs/PRD.md#L321)

## Current Status

1. Realistic historical dataset used for training and evaluation: In progress
- Evidence: `data/raw/prod_invoices_dataset.csv`, `data/raw/prod_invoice_payment_rows.csv`
- Evidence: `notebooks/test/prod_baseline_evaluation.ipynb` builds labels from real extracted data.
- Remaining: freeze one final cohort snapshot and record extraction date/window in notebook output.

2. Model comparison and model selection rationale: In progress
- Evidence: baseline notebook compares trivial baselines and logistic regression.
- Remaining: run 2 to 3 candidate models on the same split and document selection rationale.

3. Scored open-balance set with prediction and confidence: In progress
- Evidence: `data/processed/scored_predictions.csv` exists.
- Remaining: confirm this file comes from the selected model and a representative open-balance cohort.

4. UI view of a specific patient in collection workflow: Not started
- Remaining: show one patient-level screen state with prediction, confidence, and fallback behavior when no prediction exists.

5. Extension story to 30, 60, 90 day prediction and financing recommendations: In progress
- Evidence: timing bucket analysis completed in baseline notebook and 30/60/90 rates computed from real payment rows.
- Remaining: add a short product narrative for 30/60/90 extension and financing recommendations.

6. Office recommendation policy layered on prediction: In progress
- Evidence: `docs/collection-recommendation-v1.md` defines payload contract, deterministic rules, and success metrics.
- Remaining: emit recommendation in scoring output and surface one actionable recommendation card in UI.

## Done Means

This project is done when all five PRD items are complete and demo-ready, with evidence in the repo:
- fixed data snapshot and notebook outputs,
- selected model with rationale,
- scored output artifact,
- UI walkthrough for one patient,
- explicit 30/60/90 and financing extension narrative.

Recommendation bolt-on is demo-ready when:
- prediction response includes recommendation fields,
- office-facing UI shows one primary action plus fallback,
- recommendation events and payment outcomes are logged for measurement.

## Suggested Final Demo Sequence

1. Show data extract files and baseline notebook outputs.
2. Show model comparison table and selected model.
3. Show scored predictions artifact.
4. Show UI for one patient with prediction and confidence.
5. Close with extension path for 30/60/90 and financing actions.

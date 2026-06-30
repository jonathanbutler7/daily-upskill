# Success Criteria

Use this as a pass/fail scorecard.

## 1) Prediction Accuracy

- Overall ranking quality (ROC-AUC): >= 0.70 (stretch >= 0.75)
- Performance on unpaid accounts (PR-AUC): >= 25% better than a simple prevalence baseline
- When we predict "unlikely to pay," we should be right at least 65% of the time (precision >= 0.65)
- We should catch at least 40% of truly risky accounts (recall >= 0.40)

## 2) Risk Tier Quality

- Tiers: high >= 0.70, medium 0.40-0.69, low < 0.40
- Payment rate must step down cleanly: high > medium > low
- Gap between high and low tier payment rates: >= 20 percentage points
- Tier coverage: >= 85% of accounts should fall into high or low (not mostly medium)

## 3) Better Than Simple Rules

- Compare against two simple baselines:
  - Always predict the majority class
  - One-rule cutoff, for example on historical_on_time_rate
- Our model must beat both baselines on:
  - ROC-AUC
  - F1 on the unlikely-to-pay group

## 4) Data and Scoring Reliability

- Training rows with usable labels: >= 95%
- Missing values in each required feature after cleanup: <= 10%
- ZIP to Census match rate: >= 90%
- Batch scoring runtime on demo set: < 5 minutes
- Database write success for predictions: 100%

## 5) Demo Readiness Checklist

- [ ] Full run works end to end (pull, join, score, write/export)
- [ ] Output includes: will_pay_in_30_days, confidence_score, risk_band, model_version, scored_at
- [ ] Score at least 1,000 open invoices (or all available if fewer)
- [ ] UI handles both states: has_prediction=true and has_prediction=false
- [ ] Save evaluation outputs with each model: confusion matrix, ROC-AUC, precision/recall, tier payment rates

## 6) Phase Pass/Fail

This phase passes only if all five sections above pass.

## 7) Stretch Goals (PRD Alignment)

- [ ] Business impact check: estimate expected collections lift or late-payment reduction from model-guided actions.
- [ ] Model selection rationale is documented: include baseline comparison summary and why the final model was chosen.
- [ ] Extension story is demo-ready: outline concrete next step to 30/60/90-day predictions and financing recommendations.

## 8) Definitions of terms

- ROC-AUC: How well the model ranks likely payers above unlikely payers across all score cutoffs. 0.5 is random, 1.0 is perfect.
- PR-AUC: How well the model finds the unpaid/risky group when that group is smaller. Useful when classes are imbalanced.
- Precision: Of the accounts we flagged as unlikely to pay, how many were actually unlikely to pay.
- Recall: Of all truly unlikely-to-pay accounts, how many we successfully flagged.
- F1: A single score that balances precision and recall.
- Baseline: A simple reference method the model must beat.
- Majority-class baseline: Always predict whichever outcome is most common.
- Prevalence baseline: Performance you would expect from class frequency alone (for example if 30% are unpaid).
- Confusion matrix: A count table of right and wrong predictions (true/false positives and negatives).
- Risk tier: Bucket made from model score (high, medium, low) to support action decisions.
- Label: The ground-truth outcome we train on, here paid within 30 days or not.
- Label completeness: Percent of rows that have a usable label.
- ZIP to Census match rate: Percent of rows where ZIP code successfully joins to Census features.
- Batch scoring: Running predictions in a scheduled job on many accounts at once, not one request at a time.
- Upsert: Insert a new row, or update the existing row if it already exists.

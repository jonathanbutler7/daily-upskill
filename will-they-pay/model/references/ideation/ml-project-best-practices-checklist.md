# ML Project Best-Practices Checklist

This checklist maps common ML production best practices to the current state of this repository.

Status legend:
- `DONE`: present in repo.
- `PARTIAL`: present but incomplete.
- `MISSING`: not present yet.

## 1) Project Structure And Documentation

- `DONE`: product/technical docs (`docs/PRD.md`, `docs/TPD.md`).
- `DONE`: data-readiness and definition-of-done docs under `docs/ideation/`.
- `DONE`: db schema references under `docs/db-schemas/`.
- `DONE`: model workspace README (`model/README.md`).
- `DONE`: explicit active model + input/output contract (`docs/model-runtime-contract.md`).

## 2) Reproducible Environment

- `DONE`: Python training/runtime dependencies (`model/requirements.txt`).
- `DONE`: dev dependencies (`model/requirements-dev.txt`).
- `DONE`: lint/test/format defaults (`model/pyproject.toml`).
- `PARTIAL`: strict lockfile for deterministic installs (recommend `pip-compile` output committed).

## 3) Data Management

- `DONE`: clear data folder layout (`data/raw`, `data/processed`, `data/output`).
- `DONE`: scripts for extraction and preprocessing (`model/scripts/`).
- `PARTIAL`: snapshot/version manifest for training cohorts (file with extraction window, row counts, hash).
- `PARTIAL`: automated data validation checks as executable tests.

## 4) Training And Evaluation

- `DONE`: production notebook training flow (`notebooks/production/train_model_prod.ipynb`).
- `DONE`: candidate model comparison and metrics in eval summary artifact.
- `DONE`: saved feature contract (`models/feature_columns_prod.json`).
- `DONE`: saved evaluation summary (`models/eval_summary_prod.json`).
- `PARTIAL`: serialized model artifact present in repo/workspace (`models/will_they_pay_prod_v1.joblib` currently missing).

## 5) Serving And Inference Safety

- `DONE`: scoring output consumed by service (`data/output/scored_predictions_prod.csv`).
- `DONE`: service behavior for missing predictions covered by tests (per current repo memory notes).
- `PARTIAL`: schema validation on scored CSV load (dtype, required columns, null thresholds).

## 6) Testing And Quality Gates

- `DONE`: Go CI lint/test/build pipeline in `workflows/cicd.yaml`.
- `PARTIAL`: Python unit tests for scripts/feature engineering.
- `PARTIAL`: notebook smoke test in CI (execute key notebook cells or exported training script).
- `MISSING`: model quality guardrail in CI (fail if AUC drops below threshold).

## 7) Monitoring And Drift (Production-Readiness)

- `MISSING`: prediction distribution monitoring over time.
- `MISSING`: training-serving skew checks.
- `MISSING`: data drift and label drift alerts.

## 8) Security And Governance

- `DONE`: local env and large/generated artifacts ignored in git (`.gitignore`).
- `PARTIAL`: explicit PII handling policy for raw extracts (retention and access controls documented).
- `MISSING`: model card documenting intended use, limits, and fairness considerations.

## High-Impact Next Steps

1. Add deterministic lockfile generation (`pip-tools`) and commit lockfile.
2. Export and store `will_they_pay_prod_v1.joblib` in your artifact storage path used by training/release.
3. Add Python tests for `model/scripts/` transformations and label-generation logic.
4. Add CI model quality threshold check (for example, fail if ROC-AUC < 0.90).
5. Add a model card under `model/docs/` with assumptions, exclusions, and known risks.

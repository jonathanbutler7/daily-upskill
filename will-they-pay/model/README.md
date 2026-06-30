# Will They Pay — Model Workspace

This folder contains everything needed to train, evaluate, and run the Will They Pay
prediction model: data pipeline, training notebooks, scoring scripts, the model artifact,
and supporting documentation.

## Setup

```bash
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
pip install -r requirements-dev.txt
```

---

## Baseline Data

Raw data lives in `data/raw/invoice_cohorts/`. These are dated CSV snapshots pulled from
production using the scripts in `scripts/`. They are the source of record for training.

| File | Contents |
|------|----------|
| `data/raw/invoice_cohorts/invoices_*.csv` | Raw invoice rows from `payments.invoices` |
| `data/raw/invoice_cohorts/payment_rows_*.csv` | Payment log rows for those invoices |
| `data/raw/invoice_cohorts/person_invoice_history_*.csv` | Full invoice history by person |
| `data/processed/census_zip_features_model_ready.csv` | ZIP-level census features (regenerate with `make census-all`) |
| `data/external/usps-zip-codes/` | USPS ZIP code reference tables |

To refresh raw data, run the scripts in `scripts/` in order (see `scripts/README.md`).
Census data can be refreshed via `make census-all` from the repo root.

---

## Training

Training happens in two notebooks, run in order from `notebooks/production/`:

1. **`build_training_data_prod.ipynb`** — Joins invoice data with census features,
   derives behavioral features, and writes `data/processed/training_data_prod.csv`.

2. **`train_model_prod.ipynb`** — Loads the flat training table, trains and evaluates
   candidate models (Logistic Regression, Random Forest, XGBoost), and exports artifacts.

### Training Outputs

| Artifact | Path | Purpose |
|----------|------|---------|
| Model | `models/will_they_pay_prod_v1.joblib` | Trained XGBoost model (**gitignored — must regenerate**) |
| Feature contract | `models/feature_columns_prod.json` | Ordered list of 17 input features |
| Eval summary | `models/eval_summary_prod.json` | ROC-AUC, Brier score, row counts |
| Feature importance | `models/feature_importance_prod.json` | Per-feature importance breakdown |

Current model performance: **ROC-AUC 0.9469**, trained on 19,480 rows with SMOTE balancing.
See `docs/ideation/FEATURE_IMPORTANCE.md` for full breakdown.

---

## The Joblib Model

The trained model is serialized to `models/will_they_pay_prod_v1.joblib` using `joblib.dump()`.
This file is **gitignored** and must be regenerated locally by running
`notebooks/production/train_model_prod.ipynb`.

The model expects exactly 17 features in the order defined in `models/feature_columns_prod.json`.
See `docs/model-runtime-contract.md` for the full input/output spec.

---

## Running the Scoring Pipeline

### Batch (score all open invoices)

Prerequisites: `data/processed/training_data_prod.csv` and `models/will_they_pay_prod_v1.joblib`
must exist (run both training notebooks first).

```bash
# From repo root:
make score-batch

# Or directly from model/:
python scripts/score_batch.py

# Custom paths:
python scripts/score_batch.py \
    --training-data data/processed/training_data_prod.csv \
    --model models/will_they_pay_prod_v1.joblib \
    --output data/output/scored_predictions_prod.csv
```

Output: `data/output/scored_predictions_prod.csv` — this is the file the Go service reads.

### Single Record (spot-check or debug)

```bash
# From repo root:
make score-single ARGS="--amount 150 --created-day-of-week 1 --created-day-of-month 15 \
    --created-hour 10 --payment-origin-mode 2 --surchargingenabled 0 \
    --is-guardian 0 --account-age-days 365 --tenure-months 12 --is-new-patient 0 \
    --average-days-to-pay 14.5 --appointment-reliability-score 0.85 \
    --median-household-income 65000 --poverty-rate-pct 0.12 \
    --average-household-size 2.5 --bachelors-or-higher-pct 0.35 \
    --unemployment-rate-pct 0.04"

# Or with a JSON file:
python scripts/score_single.py --json-file path/to/record.json
```

---

## Notebooks

| Directory | Purpose |
|-----------|---------|
| `notebooks/production/` | Operational notebooks tied to real data and model artifacts |
| `notebooks/dev/` | Exploratory and development notebooks using synthetic or dummy data |

---

## Runtime Contract

The Go service reads from `data/output/scored_predictions_prod.csv`. It does not call the
model directly. There are two contracts to keep aligned:

1. **Model input features** — defined in `models/feature_columns_prod.json`
2. **Service I/O schema** — defined in `docs/model-runtime-contract.md`

If the scored predictions CSV is stale, rerun `make score-batch`.

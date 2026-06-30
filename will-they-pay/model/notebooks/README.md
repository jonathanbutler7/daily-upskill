# Notebooks

Notebooks are split by intent: production notebooks run against real data and produce
committed artifacts; dev notebooks are for exploration and use synthetic or dummy data.

## `production/`

Operational notebooks. Run these in order to regenerate model artifacts.

| Notebook | Purpose | Output |
|----------|---------|--------|
| `build_training_data_prod.ipynb` | Joins invoice + census data, derives behavioral features | `data/processed/training_data_prod.csv` |
| `train_model_prod.ipynb` | Trains XGBoost model, evaluates performance, exports artifacts, batch-scores open invoices | `models/will_they_pay_prod_v1.joblib`, `models/eval_summary_prod.json`, `data/output/scored_predictions_prod.csv` |
| `generate_prod_dummy_data.ipynb` | Generates synthetic production-shaped data for offline testing | dummy CSVs |

## `dev/`

Exploratory and development notebooks. Not tied to production artifacts.

| Notebook | Purpose |
|----------|---------|
| `will_they_pay_model_dummy_data.ipynb` | Full model train/eval loop on dummy data (fast iteration) |
| `build_training_data.ipynb` | Feature engineering experiments |
| `generate_dummy_data.ipynb` | Generate synthetic data for dev use |
| `prod_baseline_evaluation.ipynb` | Baseline model comparison against production data |

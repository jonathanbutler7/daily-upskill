PYTHON ?= python3
MODEL_DIR := model
MODEL_VENV_DVC := $(MODEL_DIR)/.venv/bin/dvc

CENSUS_FIELDS ?= DP03_0062E,DP03_0063E,DP03_0128PE,DP03_0005PE,DP03_0097PE,DP03_0098PE,DP04_0134E,DP04_0089E,DP02_0016E,DP02PR_0016E,DP02_0068PE,DP02PR_0068PE

.PHONY: census-raw census-model census-all score-batch score-single dvc-stages dvc-repro dvc-metrics

## Census data pipeline (run these before training)
census-raw:
	cd $(MODEL_DIR) && $(PYTHON) scripts/fetch_census_income_by_zip.py \
		--fields "$(CENSUS_FIELDS)" \
		--output data/processed/census_zip_features_raw.csv \
		--progress-every 10000

census-model:
	cd $(MODEL_DIR) && $(PYTHON) scripts/build_model_ready_census_features.py \
		--input data/processed/census_zip_features_raw.csv \
		--output data/processed/census_zip_features_model_ready.csv

census-all: census-raw census-model

## Scoring (requires models/will_they_pay_prod_v1.joblib and data/processed/training_data_prod.csv)
## Run notebooks/production/train_model_prod.ipynb to generate both prerequisites.
score-batch:
	cd $(MODEL_DIR) && $(PYTHON) scripts/score_batch.py

## Score a single record — pass features via flags or --json-file. See script for usage.
score-single:
	cd $(MODEL_DIR) && $(PYTHON) scripts/score_single.py $(ARGS)

## DVC pipeline helpers (see model/dvc.yaml)
dvc-stages:
	cd $(MODEL_DIR) && ../$(MODEL_VENV_DVC) stage list

## Reproduce one stage by name, e.g. make dvc-repro STAGE=train_model_prod
dvc-repro:
	cd $(MODEL_DIR) && ../$(MODEL_VENV_DVC) repro $(STAGE)

## Show tracked metrics, including eval_summary_prod.json
dvc-metrics:
	cd $(MODEL_DIR) && ../$(MODEL_VENV_DVC) metrics show

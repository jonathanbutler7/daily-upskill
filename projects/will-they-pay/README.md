# will-they-pay-service 
[![CI/CD](https://github.com/weave-lab/will-they-pay-service/actions/workflows/cicd.yaml/badge.svg)](https://github.com/weave-lab/will-they-pay-service/actions/workflows/cicd.yaml)

Service to help estimate whether a customer will pay.

## Model Contract

The canonical model runtime contract is documented in [model/docs/model-runtime-contract.md](model/docs/model-runtime-contract.md).

Quick summary:
- Training winner: XGBoost (see [model/models/eval_summary_prod.json](model/models/eval_summary_prod.json)).
- Runtime serving path: pre-scored CSV consumed by the Go service at [model/data/output/scored_predictions_prod.csv](model/data/output/scored_predictions_prod.csv).
- Feature inputs used for scoring: [model/models/feature_columns_prod.json](model/models/feature_columns_prod.json).
- End-to-end model workspace guide: [model/README.md](model/README.md).

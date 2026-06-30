FROM scratch

COPY will-they-pay-service /
COPY model/data/output/scored_predictions_prod.csv /model/data/output/scored_predictions_prod.csv

COPY .weave.yaml /

ENTRYPOINT ["/will-they-pay-service"]

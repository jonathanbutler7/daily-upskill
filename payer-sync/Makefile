SQLC := $(shell go env GOPATH)/bin/sqlc
GOOSE := $(shell go env GOPATH)/bin/goose

-include .env
export

DB_CONN ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: sqlc-generate migrate-up migrate-down migrate-status migrate-reset seed ingest reconcile dev processor audit-report

sqlc-generate:
	$(SQLC) generate

migrate-up:
	go run . migrate up

migrate-down:
	go run . migrate down

migrate-status:
	go run . migrate status

migrate-reset:
	go run . migrate reset

seed:
	go run . seed

ingest:
	go run . ingest

reconcile:
	go run . reconcile

# Generate test data, ingest it, then reconcile — mirrors the production cron sequence.
dev: seed ingest reconcile

processor:
	go run . processor

audit-report:
	psql -P pager=off "$(DB_CONN)" -v ON_ERROR_STOP=1 -f db/docs/audit_report.sql

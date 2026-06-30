-- +goose Up

ALTER TABLE job_runs
    DROP CONSTRAINT IF EXISTS job_runs_status_check;

ALTER TABLE job_runs
    ADD CONSTRAINT job_runs_status_check
    CHECK (status IN ('running', 'success', 'failure', 'partial'));

-- +goose Down

UPDATE job_runs SET status = 'failure' WHERE status = 'running';

ALTER TABLE job_runs
    DROP CONSTRAINT IF EXISTS job_runs_status_check;

ALTER TABLE job_runs
    ADD CONSTRAINT job_runs_status_check
    CHECK (status IN ('success', 'failure', 'partial'));

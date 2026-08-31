package ledgerstore

import (
	"context"
	"database/sql"
	"sort"
)

const defaultSettlementUpdateJobLimit = 100

type settlementUpdateJob struct {
	ID                  int64
	SettlementAccountID AccountID
	Amount              Amount
}

func ProcessSettlementUpdateJobs(ctx context.Context, db *sql.DB, limit int) (ProcessSettlementUpdateJobsResult, error) {
	if limit <= 0 {
		limit = defaultSettlementUpdateJobLimit
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ProcessSettlementUpdateJobsResult{}, err
	}
	defer tx.Rollback()

	jobs, err := lockPendingSettlementUpdateJobs(ctx, tx, limit)
	if err != nil {
		return ProcessSettlementUpdateJobsResult{}, err
	}
	if len(jobs) == 0 {
		if err := tx.Commit(); err != nil {
			return ProcessSettlementUpdateJobsResult{}, err
		}
		return ProcessSettlementUpdateJobsResult{}, nil
	}

	amountByAccount := make(map[AccountID]Amount)
	for _, job := range jobs {
		amountByAccount[job.SettlementAccountID] += job.Amount
	}

	accountIDs := make([]AccountID, 0, len(amountByAccount))
	for accountID := range amountByAccount {
		accountIDs = append(accountIDs, accountID)
	}
	sort.Slice(accountIDs, func(i, j int) bool {
		return accountIDs[i] < accountIDs[j]
	})

	for _, accountID := range accountIDs {
		if _, _, err := lockAccountForUpdate(ctx, tx, accountID); err != nil {
			return ProcessSettlementUpdateJobsResult{}, err
		}
		if err := adjustAccountBalance(ctx, tx, accountID, amountByAccount[accountID]); err != nil {
			return ProcessSettlementUpdateJobsResult{}, err
		}
	}

	for _, job := range jobs {
		if err := markSettlementUpdateJobApplied(ctx, tx, job.ID); err != nil {
			return ProcessSettlementUpdateJobsResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return ProcessSettlementUpdateJobsResult{}, err
	}

	return ProcessSettlementUpdateJobsResult{
		JobsProcessed:   len(jobs),
		AccountsUpdated: len(accountIDs),
	}, nil
}

func lockPendingSettlementUpdateJobs(ctx context.Context, tx *sql.Tx, limit int) ([]settlementUpdateJob, error) {
	const q = `
		select id, settlement_account_id, amount
		from settlement_update_jobs
		where status = 'pending'
		order by id
		limit $1
		for update skip locked;
	`

	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []settlementUpdateJob
	for rows.Next() {
		var job settlementUpdateJob
		if err := rows.Scan(&job.ID, &job.SettlementAccountID, &job.Amount); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func markSettlementUpdateJobApplied(ctx context.Context, tx *sql.Tx, jobID int64) error {
	const q = `
		update settlement_update_jobs
		set status = 'applied',
			attempts = attempts + 1,
			processed_at = now(),
			last_error = null
		where id = $1
			and status = 'pending';
	`

	_, err := tx.ExecContext(ctx, q, jobID)
	return err
}

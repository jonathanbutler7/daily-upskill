package ledgervalidator

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CheckName string

const (
	CheckTransactionBalances CheckName = "transaction_balances"
	CheckAccountBalances     CheckName = "account_balances"
	CheckTransactionShape    CheckName = "transaction_shape"
	CheckExternalTransfers   CheckName = "external_transfers"
	CheckReversals           CheckName = "reversals"
	CheckPostedOnlyState     CheckName = "posted_only_state"
)

const defaultMaxIssuesPerCheck = 100

type Options struct {
	MaxIssuesPerCheck int
}

type Result struct {
	Healthy    bool          `json:"healthy"`
	CheckedAt  time.Time     `json:"checked_at"`
	SnapshotAt time.Time     `json:"snapshot_at"`
	Summary    Summary       `json:"summary"`
	Checks     []CheckResult `json:"checks"`
	Issues     []Issue       `json:"issues"`
}

type Summary struct {
	AccountCount          int `json:"account_count"`
	TransactionCount      int `json:"transaction_count"`
	EntryCount            int `json:"entry_count"`
	ExternalTransferCount int `json:"external_transfer_count"`
	ReversalCount         int `json:"reversal_count"`
	IssueCount            int `json:"issue_count"`
}

type CheckResult struct {
	Name       CheckName `json:"name"`
	Healthy    bool      `json:"healthy"`
	IssueCount int       `json:"issue_count"`
}

type Issue struct {
	Check              CheckName `json:"check"`
	Code               string    `json:"code"`
	Message            string    `json:"message"`
	AccountID          *int64    `json:"account_id,omitempty"`
	TransactionID      *int64    `json:"transaction_id,omitempty"`
	ExternalTransferID *int64    `json:"external_transfer_id,omitempty"`
	ReversalID         *int64    `json:"reversal_id,omitempty"`
	StoredAmount       *int64    `json:"stored_amount,omitempty"`
	DerivedAmount      *int64    `json:"derived_amount,omitempty"`
	Delta              *int64    `json:"delta,omitempty"`
}

func Validate(ctx context.Context, db *sql.DB, opts Options) (Result, error) {
	checkedAt := time.Now().UTC()
	limit := opts.MaxIssuesPerCheck
	if limit <= 0 {
		limit = defaultMaxIssuesPerCheck
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly:  true,
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()

	snapshotAt, err := snapshotTime(ctx, tx)
	if err != nil {
		return Result{}, err
	}

	summary, err := loadSummary(ctx, tx)
	if err != nil {
		return Result{}, err
	}

	checks := []struct {
		name CheckName
		run  func(context.Context, *sql.Tx, int) ([]Issue, error)
	}{
		{CheckTransactionBalances, checkTransactionBalances},
		{CheckAccountBalances, checkAccountBalances},
		{CheckTransactionShape, checkTransactionShape},
		{CheckExternalTransfers, checkExternalTransfers},
		{CheckReversals, checkReversals},
		{CheckPostedOnlyState, checkPostedOnlyState},
	}

	result := Result{
		Healthy:    true,
		CheckedAt:  checkedAt,
		SnapshotAt: snapshotAt,
		Summary:    summary,
		Checks:     make([]CheckResult, 0, len(checks)),
		Issues:     make([]Issue, 0),
	}

	for _, check := range checks {
		issues, err := check.run(ctx, tx, limit)
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", check.name, err)
		}
		result.Issues = append(result.Issues, issues...)
		result.Checks = append(result.Checks, CheckResult{
			Name:       check.name,
			Healthy:    len(issues) == 0,
			IssueCount: len(issues),
		})
	}

	result.Summary.IssueCount = len(result.Issues)
	result.Healthy = result.Summary.IssueCount == 0

	if err := tx.Commit(); err != nil {
		return Result{}, err
	}

	return result, nil
}

func snapshotTime(ctx context.Context, tx *sql.Tx) (time.Time, error) {
	var snapshotAt time.Time
	err := tx.QueryRowContext(ctx, `select transaction_timestamp();`).Scan(&snapshotAt)
	return snapshotAt, err
}

func loadSummary(ctx context.Context, tx *sql.Tx) (Summary, error) {
	const q = `
		select
			(select count(*) from ledger_accounts),
			(select count(*) from ledger_transactions),
			(select count(*) from ledger_entries),
			(select count(*) from external_transfers),
			(select count(*) from ledger_reversals);
	`

	var summary Summary
	err := tx.QueryRowContext(ctx, q).Scan(
		&summary.AccountCount,
		&summary.TransactionCount,
		&summary.EntryCount,
		&summary.ExternalTransferCount,
		&summary.ReversalCount,
	)
	return summary, err
}

func checkTransactionBalances(ctx context.Context, tx *sql.Tx, limit int) ([]Issue, error) {
	const q = `
		select
			lt.id,
			coalesce(sum(le.amount), 0) as entry_sum,
			count(le.id) as entry_count
		from ledger_transactions lt
		left join ledger_entries le on le.transaction_id = lt.id
			and le.archived = false
		where lt.status = 'posted'
		group by lt.id
		having coalesce(sum(le.amount), 0) <> 0
			or count(le.id) = 0
		order by lt.id
		limit $1;
	`

	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var transactionID int64
		var entrySum int64
		var entryCount int64
		if err := rows.Scan(&transactionID, &entrySum, &entryCount); err != nil {
			return nil, err
		}

		issues = append(issues, Issue{
			Check:         CheckTransactionBalances,
			Code:          "transaction_not_balanced",
			Message:       fmt.Sprintf("posted transaction %d has %d active entries with sum %d", transactionID, entryCount, entrySum),
			TransactionID: int64Ptr(transactionID),
			Delta:         int64Ptr(entrySum),
		})
	}

	return issues, rows.Err()
}

func checkAccountBalances(ctx context.Context, tx *sql.Tx, limit int) ([]Issue, error) {
	const q = `
		select
			la.id,
			la.balance as stored_balance,
			coalesce(sum(case
				when lt.status = 'posted'
					and le.archived = false then le.amount
				else 0
			end), 0) as derived_balance
		from ledger_accounts la
		left join ledger_entries le on le.account_id = la.id
		left join ledger_transactions lt on lt.id = le.transaction_id
		group by la.id, la.balance
		having la.balance <> coalesce(sum(case
			when lt.status = 'posted'
				and le.archived = false then le.amount
			else 0
		end), 0)
		order by la.id
		limit $1;
	`

	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var accountID int64
		var storedBalance int64
		var derivedBalance int64
		if err := rows.Scan(&accountID, &storedBalance, &derivedBalance); err != nil {
			return nil, err
		}

		delta := storedBalance - derivedBalance
		issues = append(issues, Issue{
			Check:         CheckAccountBalances,
			Code:          "account_balance_mismatch",
			Message:       fmt.Sprintf("account %d stored balance %d does not match derived balance %d", accountID, storedBalance, derivedBalance),
			AccountID:     int64Ptr(accountID),
			StoredAmount:  int64Ptr(storedBalance),
			DerivedAmount: int64Ptr(derivedBalance),
			Delta:         int64Ptr(delta),
		})
	}

	return issues, rows.Err()
}

func checkTransactionShape(ctx context.Context, tx *sql.Tx, limit int) ([]Issue, error) {
	const q = `
		select
			lt.id,
			lt.amount,
			count(le.id) as entry_count
		from ledger_transactions lt
		left join ledger_entries le on le.transaction_id = lt.id
			and le.archived = false
		where lt.status = 'posted'
		group by lt.id
		having count(le.id) <> 2
			or sum(case when le.account_id = lt.from_account_id
				and le.amount = -lt.amount then 1 else 0 end) <> 1
			or sum(case when le.account_id = lt.to_account_id
				and le.amount = lt.amount then 1 else 0 end) <> 1
		order by lt.id
		limit $1;
	`

	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var transactionID int64
		var amount int64
		var entryCount int64
		if err := rows.Scan(&transactionID, &amount, &entryCount); err != nil {
			return nil, err
		}

		issues = append(issues, Issue{
			Check:         CheckTransactionShape,
			Code:          "transaction_entry_shape_mismatch",
			Message:       fmt.Sprintf("posted transaction %d has %d active entries that do not match its from/to amount %d", transactionID, entryCount, amount),
			TransactionID: int64Ptr(transactionID),
			StoredAmount:  int64Ptr(amount),
		})
	}

	return issues, rows.Err()
}

func checkExternalTransfers(ctx context.Context, tx *sql.Tx, limit int) ([]Issue, error) {
	const q = `
		select et.id, et.ledger_transaction_id
		from external_transfers et
		left join ledger_transactions lt on lt.id = et.ledger_transaction_id
		left join ledger_accounts cs
			on cs.name = 'Cash Settlement'
			and cs.currency_code = et.currency_code
		where et.status = 'posted'
			and (
				lt.id is null
				or cs.id is null
				or et.completed_at is null
				or lt.status <> 'posted'
				or lt.type <> et.direction
				or lt.amount <> et.amount
				or lt.currency_code <> et.currency_code
				or (et.direction = 'deposit'
					and (lt.from_account_id <> cs.id
						or lt.to_account_id <> et.user_account_id))
				or (et.direction = 'withdrawal'
					and (lt.from_account_id <> et.user_account_id
						or lt.to_account_id <> cs.id))
			)
		order by et.id
		limit $1;
	`

	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var externalTransferID int64
		var transactionID sql.NullInt64
		if err := rows.Scan(&externalTransferID, &transactionID); err != nil {
			return nil, err
		}

		issue := Issue{
			Check:              CheckExternalTransfers,
			Code:               "external_transfer_mismatch",
			Message:            fmt.Sprintf("external transfer %d does not match its linked posted ledger transaction", externalTransferID),
			ExternalTransferID: int64Ptr(externalTransferID),
		}
		if transactionID.Valid {
			issue.TransactionID = int64Ptr(transactionID.Int64)
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

func checkReversals(ctx context.Context, tx *sql.Tx, limit int) ([]Issue, error) {
	const q = `
		with original_entries as (
			select
				lr.id as reversal_id,
				lr.reversal_transaction_id,
				le.account_id,
				sum(le.amount) as amount
			from ledger_reversals lr
			join ledger_entries le on le.transaction_id = lr.original_transaction_id
			where le.archived = false
			group by lr.id, lr.reversal_transaction_id, le.account_id
		),
		reversal_entries as (
			select
				lr.id as reversal_id,
				lr.reversal_transaction_id,
				le.account_id,
				sum(le.amount) as amount
			from ledger_reversals lr
			join ledger_entries le on le.transaction_id = lr.reversal_transaction_id
			where le.archived = false
			group by lr.id, lr.reversal_transaction_id, le.account_id
		),
		entry_mismatches as (
			select
				coalesce(o.reversal_id, r.reversal_id) as reversal_id,
				coalesce(o.reversal_transaction_id, r.reversal_transaction_id) as reversal_transaction_id,
				coalesce(o.account_id, r.account_id) as account_id,
				coalesce(o.amount, 0) as original_amount,
				coalesce(r.amount, 0) as reversal_amount
			from original_entries o
			full outer join reversal_entries r
				on r.reversal_id = o.reversal_id
				and r.account_id = o.account_id
			where coalesce(o.amount, 0) + coalesce(r.amount, 0) <> 0
		),
		header_mismatches as (
			select
				lr.id as reversal_id,
				lr.reversal_transaction_id,
				null::bigint as account_id,
				original.amount as original_amount,
				reversal.amount as reversal_amount
			from ledger_reversals lr
			join ledger_transactions original on original.id = lr.original_transaction_id
			join ledger_transactions reversal on reversal.id = lr.reversal_transaction_id
			where reversal.type <> 'reversal'
				or reversal.status <> 'posted'
				or reversal.amount <> original.amount
				or reversal.currency_code <> original.currency_code
		)
		select
			reversal_id,
			reversal_transaction_id,
			account_id,
			original_amount,
			reversal_amount
		from (
			select * from entry_mismatches
			union all
			select * from header_mismatches
		) mismatches
		order by reversal_id, reversal_transaction_id, account_id nulls first
		limit $1;
	`

	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var reversalID int64
		var transactionID int64
		var accountID sql.NullInt64
		var originalAmount int64
		var reversalAmount int64
		if err := rows.Scan(&reversalID, &transactionID, &accountID, &originalAmount, &reversalAmount); err != nil {
			return nil, err
		}

		issue := Issue{
			Check:         CheckReversals,
			Code:          "reversal_mismatch",
			Message:       fmt.Sprintf("reversal %d does not fully negate transaction state", reversalID),
			ReversalID:    int64Ptr(reversalID),
			TransactionID: int64Ptr(transactionID),
			StoredAmount:  int64Ptr(reversalAmount),
			DerivedAmount: int64Ptr(-originalAmount),
			Delta:         int64Ptr(originalAmount + reversalAmount),
		}
		if accountID.Valid {
			issue.AccountID = int64Ptr(accountID.Int64)
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

func checkPostedOnlyState(ctx context.Context, tx *sql.Tx, limit int) ([]Issue, error) {
	const q = `
		select id
		from ledger_transactions
		where status <> 'posted'
		order by id
		limit $1;
	`

	rows, err := tx.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var transactionID int64
		if err := rows.Scan(&transactionID); err != nil {
			return nil, err
		}

		issues = append(issues, Issue{
			Check:         CheckPostedOnlyState,
			Code:          "non_posted_transaction",
			Message:       fmt.Sprintf("transaction %d is not posted in the current posted-only ledger", transactionID),
			TransactionID: int64Ptr(transactionID),
		})
	}

	return issues, rows.Err()
}

func int64Ptr(value int64) *int64 {
	return &value
}

package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
)

// small deposit_funds SQL helper functions

// Validate the to_account_id exists and lock row
func lockToAccountCurrencyForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	toAccountID AccountID,
) (CurrencyCode, error) {
	const q = `
		select currency_code
		from ledger_accounts
		where id = $1
		for update;
	`

	var currencyCode string
	err := tx.QueryRowContext(ctx, q, toAccountID).Scan(&currencyCode)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrToAccountNotFound
	}
	if err != nil {
		return "", err
	}
	return CurrencyCode(currencyCode), nil
}

// Locate the internal settlement account
func lockCashSettlementAccountForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	currencyCode CurrencyCode,
) (AccountID, error) {
	const q = `
		select id
		from ledger_accounts
		where name = 'Cash Settlement'
			and currency_code = $1
		for update;
	`

	var fundingAccountID int64
	err := tx.QueryRowContext(ctx, q, currencyCode).Scan(&fundingAccountID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrCashSettlementAccountNotFound
	}
	if err != nil {
		return 0, err
	}

	return AccountID(fundingAccountID), nil
}

// Check same idempotency request
// caller should treat rows not found as a chance to keep going, the request is valid
func checkSameIdempotencyRequest(
	ctx context.Context,
	tx *sql.Tx,
	idempotencyKey IdempotencyKey,
	fundingAccountID AccountID,
	toAccountID AccountID,
	transferAmount Amount,
	toCurrency CurrencyCode,
) (TransactionID, error) {
	const q = `
		select id
		from ledger_transactions lt
		where lt.idempotency_key = $1
			and lt.type = 'deposit'
			and lt.from_account_id = $2
			and lt.to_account_id = $3
			and lt.amount = $4
			and lt.currency_code = $5;
	`

	var transactionID int64
	err := tx.QueryRowContext(
		ctx,
		q,
		idempotencyKey,
		fundingAccountID,
		toAccountID,
		transferAmount,
		toCurrency,
	).Scan(&transactionID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoRowsFound
	}
	if err != nil {
		return 0, err
	}
	return TransactionID(transactionID), nil
}

// Check same idempotency conflict
// caller should treat rows not found as a chance to keep going, the request is valid
func checkIdempotencyConflict(
	ctx context.Context,
	tx *sql.Tx,
	idempotencyKey IdempotencyKey,
) (TransactionID, error) {
	const q = `
		select id
		from ledger_transactions lt
		where lt.idempotency_key = $1;
	`
	var transactionID int64
	err := tx.QueryRowContext(
		ctx,
		q,
		idempotencyKey,
	).Scan(&transactionID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoRowsFound
	}
	if err != nil {
		return 0, err
	}
	return TransactionID(transactionID), nil
}

// Insert ledger transaction
func insertLedgerTransaction(
	ctx context.Context,
	tx *sql.Tx,
	idempotencyKey IdempotencyKey,
	fundingAccountID AccountID,
	toAccountID AccountID,
	transferAmount Amount,
	toCurrency CurrencyCode,
) (TransactionID, error) {
	const q = `
		insert into ledger_transactions (
			type,
			idempotency_key,
			from_account_id,
			to_account_id,
			amount,
			currency_code
		)
		values (
			'deposit',
			$1,
			$2,
			$3,
			$4,
			$5
		)
		returning id
	`
	var transactionID int64

	err := tx.QueryRowContext(
		ctx,
		q,
		idempotencyKey,
		fundingAccountID,
		toAccountID,
		transferAmount,
		toCurrency,
	).Scan(&transactionID)
	if err != nil {
		return 0, err
	}
	return TransactionID(transactionID), nil
}

// Insert ledger entries
func insertLedgerEntries(
	ctx context.Context,
	tx *sql.Tx,
	transactionID TransactionID,
	transferAmount Amount,
	fundingAccountID AccountID,
	toAccountID AccountID,
) error {
	const q = `
		insert into ledger_entries (transaction_id, account_id, amount)
    values
        ($1, $2, $3),
        ($4, $5, $6);
	`
	_, err := tx.ExecContext(
		ctx,
		q,
		transactionID,
		fundingAccountID,
		-transferAmount,
		transactionID,
		toAccountID,
		transferAmount,
	)
	if err != nil {
		return err
	}
	return nil
}

// Update balances
func updateBalances(
	ctx context.Context,
	tx *sql.Tx,
	transferAmount Amount,
	fundingAccountID AccountID,
	toAccountID AccountID,
) error {
	const q = `
		update ledger_accounts
		set balance = balance - $1
		where id = $2;

		update ledger_accounts
		set balance = balance + $3
		where id = $4;
	`

	_, err := tx.ExecContext(
		ctx,
		q,
		transferAmount,
		fundingAccountID,
		transferAmount,
		toAccountID,
	)
	if err != nil {
		return err
	}
	return nil
}

func insertExternalTransfers(
	ctx context.Context,
	tx *sql.Tx,
	direction ExternalTransferDirection,
	rail PaymentRail,
	status ExternalTransferStatus,
	externalReference ExternalReference,
	toAccountID AccountID,
	newTransactionID TransactionID,
	transferAmount Amount,
	toCurrency CurrencyCode,
) error {
	const q = `
		insert into external_transfers (
			direction, 
			rail, 
			status, 
			external_reference, 
			user_account_id, 
			ledger_transaction_id, 
			amount,
			currency_code,
			completed_at
		)
		values (
			$1, 
			$2, 
			$3,
			$4, 
			$5, 
			$6, 
			$7,
			$8,
			now()
		);
	`

	_, err := tx.ExecContext(
		ctx,
		q,
		direction,
		rail,
		status,
		externalReference,
		toAccountID,
		newTransactionID,
		transferAmount,
		toCurrency,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoRowsFound
	}
	if err != nil {
		return err
	}
	return nil
}

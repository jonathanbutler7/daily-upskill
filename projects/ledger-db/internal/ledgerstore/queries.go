package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
)

/****************

PostTransfer SQL helper functions

*****************/

// 000 Look up and lock the from account
func lockFromAccount(ctx context.Context, tx *sql.Tx, fromAccountID AccountID) (Amount, CurrencyCode, error) {
	const q = `
		select balance, currency_code
		from ledger_accounts
		where id = $1
		for update;
	`

	var balance int64
	var currencyCode string
	err := tx.QueryRowContext(ctx, q, fromAccountID).Scan(&balance, &currencyCode)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrFromAccountNotFound
	}
	if err != nil {
		return 0, "", err
	}
	return Amount(balance), CurrencyCode(currencyCode), nil
}

// 001 Look up and lock the to account
func lockToAccount(ctx context.Context, tx *sql.Tx, toAccountID AccountID) (CurrencyCode, error) {
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

// 002 Check currencies match
func checkCurrencyMatch(fromCurrency, toCurrency CurrencyCode) error {
	if fromCurrency != toCurrency {
		return ErrCurrencyMismatch
	}
	return nil
}

// 003 If this exact request already posted, return its transaction id
func checkIdempotencyRequest(
	ctx context.Context,
	tx *sql.Tx,
	idempotencyKey IdempotencyKey,
	fromAccountID AccountID,
	toAccountID AccountID,
	transferAmount Amount,
	fromCurrency CurrencyCode,
) (TransactionID, error) {
	const q = `
		select id
		from ledger_transactions lt
		where lt.idempotency_key = $1
			and lt.from_account_id = $2
			and lt.to_account_id = $3
			and lt.amount = $4
			and lt.currency_code = $5;
	`
	var existingTransactionID int64
	err := tx.QueryRowContext(
		ctx,
		q,
		idempotencyKey,
		fromAccountID,
		toAccountID,
		transferAmount,
		fromCurrency,
	).Scan(&existingTransactionID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoRowsFound
	}
	if err != nil {
		return 0, err
	}
	return TransactionID(existingTransactionID), nil
}

// 004 If the key exists for a different request, reject it
func checkIdempotencyConflict2(
	ctx context.Context,
	tx *sql.Tx,
	idempotencyKey IdempotencyKey,
	fromAccountID AccountID,
	toAccountID AccountID,
	transferAmount Amount,
	fromCurrency CurrencyCode,
) (TransactionID, error) {
	const q = `
		select id
		from ledger_transactions lt
		where lt.idempotency_key = $1
			and not (
				lt.from_account_id = $2
				and lt.to_account_id = $3
				and lt.amount = $4
				and lt.currency_code = $5
			);
	`
	var transactionID int64
	err := tx.QueryRowContext(
		ctx,
		q,
		idempotencyKey,
		fromAccountID,
		toAccountID,
		transferAmount,
		fromCurrency,
	).Scan(&transactionID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoRowsFound
	}
	if err != nil {
		return 0, err
	}
	return TransactionID(transactionID), nil
}

// pg sleep allows us to test concurrent transaction and make sure
// race conditions are handled correctly by idempotency
// perform pg_sleep(5);

// 005 Check balance
func checkBalance(fromBalance, transferAmount Amount) error {
	if fromBalance < transferAmount {
		return ErrInsufficientFunds
	}
	return nil
}

// 006 Insert transaction
func insertTransaction(
	ctx context.Context,
	tx *sql.Tx,
	idempotencyKey IdempotencyKey,
	fromAccountID AccountID,
	toAccountID AccountID,
	transferAmount Amount,
	fromCurrency CurrencyCode,
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
			'transfer',
			$1,
			$2,
			$3,
			$4,
			$5
		)
		on conflict (idempotency_key) do nothing
		returning id;
	`

	var transactionID int64
	err := tx.QueryRowContext(
		ctx,
		q,
		idempotencyKey,
		fromAccountID,
		toAccountID,
		transferAmount,
		fromCurrency,
	).Scan(&transactionID)
	if errors.Is(err, sql.ErrNoRows) {
		existingTransactionID, lookupErr := checkIdempotencyRequest(
			ctx,
			tx,
			idempotencyKey,
			fromAccountID,
			toAccountID,
			transferAmount,
			fromCurrency,
		)
		if lookupErr == nil {
			return existingTransactionID, nil
		}
		if !errors.Is(lookupErr, ErrNoRowsFound) {
			return 0, lookupErr
		}
		return 0, ErrIdempotencyConflict
	}
	if err != nil {
		return 0, err
	}
	return TransactionID(transactionID), nil
}

// 007 Insert entries
func insertEntries(
	ctx context.Context,
	tx *sql.Tx,
	transactionID TransactionID,
	transferAmount Amount,
	fromAccountID AccountID,
	toAccountID AccountID,
) error {
	const q = `
		insert into ledger_entries (transaction_id, account_id, amount)
		values
			($1, $2, $3),
			($1, $4, $5);
	`
	_, err := tx.ExecContext(
		ctx,
		q,
		transactionID,
		fromAccountID,
		-transferAmount,
		toAccountID,
		transferAmount,
	)
	return err
}

// 008 Update balances
func updateTransferBalances(
	ctx context.Context,
	tx *sql.Tx,
	transferAmount Amount,
	fromAccountID AccountID,
	toAccountID AccountID,
) error {
	const debitFromAccount = `
		update ledger_accounts
		set balance = balance - $1
		where id = $2;
	`

	_, err := tx.ExecContext(ctx, debitFromAccount, transferAmount, fromAccountID)
	if err != nil {
		return err
	}

	const creditToAccount = `
		update ledger_accounts
		set balance = balance + $1
		where id = $2;
	`

	_, err = tx.ExecContext(ctx, creditToAccount, transferAmount, toAccountID)
	return err
}

func verifyTransactionBalances(ctx context.Context, tx *sql.Tx, transactionID TransactionID) error {
	const q = `
		select coalesce(sum(amount), 0)
		from ledger_entries
		where transaction_id = $1;
	`
	var sum int64
	err := tx.QueryRowContext(ctx, q, transactionID).Scan(&sum)
	if err != nil {
		return err
	}
	if sum != 0 {
		return ErrTransactionNotBalanced
	}
	return nil
}

/****************

DepositFunds SQL helper functions

*****************/

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
	const debitFundingAccount = `
		update ledger_accounts
		set balance = balance - $1
		where id = $2;
	`

	_, err := tx.ExecContext(
		ctx,
		debitFundingAccount,
		transferAmount,
		fundingAccountID,
	)
	if err != nil {
		return err
	}

	const creditToAccount = `
		update ledger_accounts
		set balance = balance + $1
		where id = $2;
	`

	_, err = tx.ExecContext(
		ctx,
		creditToAccount,
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

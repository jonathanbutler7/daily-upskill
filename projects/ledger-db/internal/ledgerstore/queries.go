package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
)

/****************

PostTransfer SQL helper functions

*****************/

// 0 Look up and lock an account
func lockAccountForUpdate(ctx context.Context, tx *sql.Tx, accountID AccountID) (Amount, CurrencyCode, error) {
	// jan (new employee at weave) suggested this as
	// an alternative to row locks. i am reading it now
	// and realizing i don't totally understand the intent.
	// i'll either get an understanding about it from an
	// agent or just ask him some time.
	//
	// update accounts
	// set balance = balance - 50
	// where id = ___
	// and balance > 50
	const q = `
		select balance, currency_code
		from ledger_accounts
		where id = $1
		for update;
	`

	var balance int64
	var currencyCode string
	err := tx.QueryRowContext(ctx, q, accountID).Scan(&balance, &currencyCode)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrNoRowsFound
	}
	if err != nil {
		return 0, "", err
	}
	return Amount(balance), CurrencyCode(currencyCode), nil
}

// 2 Check currencies match
func checkCurrencyMatch(fromCurrency, toCurrency CurrencyCode) error {
	if fromCurrency != toCurrency {
		return ErrCurrencyMismatch
	}
	return nil
}

// 3 If this exact request already posted, return its transaction id
func findSameLedgerTransaction(
	ctx context.Context,
	tx *sql.Tx,
	transactionType LedgerTransactionType,
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
			and lt.type = $2
			and lt.from_account_id = $3
			and lt.to_account_id = $4
			and lt.amount = $5
			and lt.currency_code = $6;
	`
	var existingTransactionID int64
	err := tx.QueryRowContext(
		ctx,
		q,
		idempotencyKey,
		transactionType,
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

// 4 If the key exists for any request, return its transaction id
func findLedgerTransactionByIdempotencyKey(
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

// pg sleep allows us to test concurrent transaction and make sure
// race conditions are handled correctly by idempotency
// perform pg_sleep(5);

// 5 Check balance
func checkBalance(fromBalance, transferAmount Amount) error {
	if fromBalance < transferAmount {
		return ErrInsufficientFunds
	}
	return nil
}

// 6 Insert transaction
func insertLedgerTransaction(
	ctx context.Context,
	tx *sql.Tx,
	transactionType LedgerTransactionType,
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
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		)
		on conflict (idempotency_key) do nothing
		returning id;
	`

	var transactionID int64
	err := tx.QueryRowContext(
		ctx,
		q,
		transactionType,
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

// 8 Update account balance
func adjustAccountBalance(
	ctx context.Context,
	tx *sql.Tx,
	accountID AccountID,
	amountDelta Amount,
	entryDirection EntryDirection,
) error {
	if entryDirection != EntryDirectionCredit && entryDirection != EntryDirectionDebit {
		return ErrMustBeWithdrawalOrDeposit
	}

	const q = `
		update ledger_accounts
		set
			balance = case
				when normal_balance = $3 then balance + $1
				else balance - $1
			end,
			lock_version = lock_version + 1
		where id = $2
			and (
				normal_balance = $3
				or balance >= $1
			);
	`

	result, err := tx.ExecContext(ctx, q, amountDelta, accountID, entryDirection)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrInsufficientFunds
	}

	return nil
}

func verifyTransactionBalances(ctx context.Context, tx *sql.Tx, transactionID TransactionID) error {
	const q = `
		select
			coalesce(sum(case when direction = 'debit' then amount else 0 end), 0) as debit_total,
			coalesce(sum(case when direction = 'credit' then amount else 0 end), 0) as credit_total
		from ledger_entries
		where transaction_id = $1;
	`

	var debitTotal int64
	var creditTotal int64
	err := tx.QueryRowContext(ctx, q, transactionID).Scan(&debitTotal, &creditTotal)
	if err != nil {
		return err
	}
	if debitTotal != creditTotal {
		return ErrTransactionNotBalanced
	}
	return nil
}

// Validate the to_account_id exists and lock row
func lockToAccountCurrencyForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	toAccountID AccountID,
) (CurrencyCode, error) {
	_, currencyCode, err := lockAccountForUpdate(ctx, tx, toAccountID)
	if errors.Is(err, ErrNoRowsFound) {
		return "", ErrToAccountNotFound
	}
	if err != nil {
		return "", err
	}
	return CurrencyCode(currencyCode), nil
}

func getAccountById(
	ctx context.Context,
	tx *sql.Tx,
	accountId AccountID,
) (Account, error) {
	const q = `
		select
			id,
			name,
			description,
			currency_code,
			normal_balance,
			ledgerable_type,
			lock_version,
			balance,
			created_at
		from ledger_accounts
		where id = $1;
	`

	var account Account
	err := tx.QueryRowContext(ctx, q, accountId).Scan(
		&account.ID,
		&account.Name,
		&account.Description,
		&account.CurrencyCode,
		&account.NormalBalance,
		&account.LedgerableType,
		&account.LockVersion,
		&account.Balance,
		&account.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNoRowsFound
	}
	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func getSettlementAccountId(
	ctx context.Context,
	tx *sql.Tx,
	settlementAccountID AccountID,
	currencyCode CurrencyCode,
) (AccountID, error) {
	if settlementAccountID == 0 {
		return getCashSettlementAccountId(ctx, tx, currencyCode)
	}

	account, err := getAccountById(ctx, tx, settlementAccountID)
	if errors.Is(err, ErrNoRowsFound) {
		return 0, ErrCashSettlementAccountNotFound
	}
	if err != nil {
		return 0, err
	}
	if account.CurrencyCode != currencyCode {
		return 0, ErrCurrencyMismatch
	}

	return account.ID, nil
}

func getCashSettlementAccountId(
	ctx context.Context,
	tx *sql.Tx,
	currencyCode CurrencyCode,
) (AccountID, error) {
	const q = `
		select id
		from ledger_accounts
		where name = 'Cash Settlement'
			and currency_code = $1;
	`
	var accountID int64
	err := tx.QueryRowContext(ctx, q, currencyCode).Scan(&accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrCashSettlementAccountNotFound
	}
	if err != nil {
		return 0, err
	}

	return AccountID(accountID), nil
}

func insertLedgerEntry(
	ctx context.Context,
	tx *sql.Tx,
	transactionID TransactionID,
	entry LedgerEntryInput,
) error {
	if entry.Amount <= 0 {
		return ErrAmountGreaterThanZero
	}

	const q = `
		insert into ledger_entries (transaction_id, account_id, amount, direction)
		values ($1, $2, $3, $4);
	`

	if _, err := tx.ExecContext(ctx, q, transactionID, entry.AccountID, entry.Amount, entry.Direction); err != nil {
		return err
	}

	return nil
}

func insertExternalTransfer(
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

func insertOutboxEvent(
	ctx context.Context,
	tx *sql.Tx,
	sourceType string,
	sourceId int64,
	eventType string,
	payload any,
) error {
	const q = `
		insert into outbox_events (
			occurred_at,
			source_type,
			source_id,
			event_type,
			payload
		)
		values (
			now(),
			$1,
			$2,
			$3,
			$4
		)
	`
	_, err := tx.ExecContext(ctx, q, sourceType, sourceId, eventType, payload)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoRowsFound
	}
	if err != nil {
		return err
	}
	return nil
}

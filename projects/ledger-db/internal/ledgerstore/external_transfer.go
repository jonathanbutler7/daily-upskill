package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func PostExternalTransfer(ctx context.Context, db *sql.DB, cmd PostExternalTransferCommand) (TransactionID, error) {
	if cmd.TransferAmount <= 0 {
		return 0, ErrAmountGreaterThanZero
	}
	if strings.TrimSpace(string(cmd.ExternalReference)) == "" {
		return 0, ErrExternalReferenceRequired
	}

	isDeposit := cmd.ExternalTransferDirection == ExternalTransferDirectionDeposit
	isWithdrawal := cmd.ExternalTransferDirection == ExternalTransferDirectionWithdrawal

	if !isDeposit && !isWithdrawal {
		return 0, ErrMustBeWithdrawalOrDeposit
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	toAccountCurrency, err := lockToAccountCurrencyForUpdate(ctx, tx, cmd.UserAccountID)
	if err != nil {
		return 0, err
	}

	cashSettlementAccountId, err := lockCashSettlementAccountForUpdate(ctx, tx, toAccountCurrency)
	if err != nil {
		return 0, err
	}

	var fromAccountID AccountID
	var toAccountID AccountID

	if cmd.ExternalTransferDirection == ExternalTransferDirectionDeposit {
		fromAccountID = cashSettlementAccountId
		toAccountID = cmd.UserAccountID
	}
	if cmd.ExternalTransferDirection == ExternalTransferDirectionWithdrawal {
		fromAccountID = cmd.UserAccountID
		toAccountID = cashSettlementAccountId
	}

	transactionID, err := findSameLedgerTransaction(ctx, tx,
		LedgerTransactionType(cmd.ExternalTransferDirection),
		cmd.IdempotencyKey,
		fromAccountID,
		toAccountID,
		cmd.TransferAmount,
		toAccountCurrency,
	)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if transactionID != 0 {
		return TransactionID(transactionID), nil
	}

	conflictingTransactionID, err := findLedgerTransactionByIdempotencyKey(ctx, tx, cmd.IdempotencyKey)
	if err != nil && !errors.Is(err, ErrNoRowsFound) {
		return 0, err
	}
	if conflictingTransactionID != 0 {
		return 0, ErrIdempotencyConflict
	}

	if cmd.ExternalTransferDirection == ExternalTransferDirectionWithdrawal {
		balance, _, err := lockFromAccount(ctx, tx, cmd.UserAccountID)
		if err != nil {
			return 0, err
		}
		err = checkBalance(balance, cmd.TransferAmount)
		if err != nil {
			return 0, err
		}
	}

	transactionID, err = insertLedgerTransaction(
		ctx,
		tx,
		LedgerTransactionType(cmd.ExternalTransferDirection),
		cmd.IdempotencyKey,
		fromAccountID,
		toAccountID,
		cmd.TransferAmount,
		toAccountCurrency,
	)
	if errors.Is(err, ErrNoRowsFound) {
		transactionID, err = findSameLedgerTransaction(
			ctx,
			tx,
			LedgerTransactionType(cmd.ExternalTransferDirection),
			cmd.IdempotencyKey,
			fromAccountID,
			toAccountID,
			cmd.TransferAmount,
			toAccountCurrency,
		)
		if err == nil {
			return transactionID, nil
		}
		if errors.Is(err, ErrNoRowsFound) {
			return 0, ErrIdempotencyConflict
		}
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	if err := insertLedgerEntries(ctx, tx, transactionID, cmd.TransferAmount, fromAccountID, toAccountID); err != nil {
		return 0, err
	}

	err = verifyTransactionBalances(ctx, tx, transactionID)
	if err != nil {
		return 0, err
	}

	if err := adjustAccountBalance(ctx, tx, fromAccountID, -cmd.TransferAmount); err != nil {
		return 0, err
	}
	if err := adjustAccountBalance(ctx, tx, toAccountID, cmd.TransferAmount); err != nil {
		return 0, err
	}

	if err := insertExternalTransfers(
		ctx,
		tx,
		cmd.ExternalTransferDirection,
		cmd.Rail,
		ExternalTransferStatusPosted,
		cmd.ExternalReference,
		// external_transfers.user_account_id always
		// points to the user ledger account. fromAccountID
		// or toAccountID may point to Cash Settlement
		// depending on direction. in this case we always
		// want the user account id.
		cmd.UserAccountID,
		transactionID,
		cmd.TransferAmount,
		toAccountCurrency,
	); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return TransactionID(transactionID), nil
}

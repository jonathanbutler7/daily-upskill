package ledgerstore

import (
	"context"
	"database/sql"
	"errors"
)

// A function to handle both depositing funds into the ledger-db
// account as well as withdrawing funds from ledger-db into a
// separate payment rail.
//
// PostExternalTransfer requires the caller to tell in the direction
// the money will move, and then it handles assigning the values
// and transferring money accordingly.
func PostExternalTransfer(ctx context.Context, db *sql.DB, cmd PostExternalTransferCommand) (TransactionID, error) {
	isDeposit := cmd.ExternalTransferDirection == ExternalTransferDirectionDeposit
	isWithdrawal := cmd.ExternalTransferDirection == ExternalTransferDirectionWithdrawal

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	userAccount, err := getAccountById(ctx, tx, cmd.UserAccountID)
	if errors.Is(err, ErrNoRowsFound) {
		return 0, ErrToAccountNotFound
	}
	if err != nil {
		return 0, err
	}

	toAccountCurrency := userAccount.CurrencyCode

	cashSettlementAccountId, err := getSettlementAccountId(ctx, tx, cmd.SettlementAccountID, toAccountCurrency)
	if err != nil {
		return 0, err
	}

	var fromAccountID AccountID
	var toAccountID AccountID

	if isDeposit {
		fromAccountID = cashSettlementAccountId
		toAccountID = cmd.UserAccountID
	}
	if isWithdrawal {
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

	entries := []LedgerEntryInput{
		{AccountID: fromAccountID, Amount: cmd.TransferAmount, Direction: EntryDirectionDebit},
		{AccountID: toAccountID, Amount: cmd.TransferAmount, Direction: EntryDirectionCredit},
	}
	for _, entry := range entries {
		if err := insertLedgerEntry(ctx, tx, transactionID, entry); err != nil {
			return 0, err
		}
	}

	err = verifyTransactionBalances(ctx, tx, transactionID)
	if err != nil {
		return 0, err
	}

	if err := adjustAccountBalance(ctx, tx, fromAccountID, cmd.TransferAmount, EntryDirectionDebit); err != nil {
		return 0, err
	}
	if err := adjustAccountBalance(ctx, tx, toAccountID, cmd.TransferAmount, EntryDirectionCredit); err != nil {
		return 0, err
	}

	if err := insertExternalTransfer(
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


	if err := insertOutboxEvent(
		ctx, tx,
		"ledger_transaction",
		int64(transactionID),
		"ledger.transaction.posted",
		"{\"payload\":\"payload\"}",
	); err != nil {
		return 0, err
	}


	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return TransactionID(transactionID), nil
}

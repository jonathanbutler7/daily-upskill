package ledgerstore

import "time"

// Named types for values that cross ledgerstore helper boundaries.
type AccountID int64
type TransactionID int64
type ReversalID int64
type Amount int64
type CurrencyCode string
type IdempotencyKey string
type PaymentRail string
type ExternalReference string
type ExternalTransferDirection string
type ExternalTransferStatus string
type LedgerTransactionType string
type Reason string
type NormalBalance string
type EntryDirection string
type LedgerableType string

const (
	LedgerTransactionTypeTransfer LedgerTransactionType = "transfer"
	LedgerTransactionTypeDeposit  LedgerTransactionType = "deposit"
	LedgerTransactionTypeReversal LedgerTransactionType = "reversal"

	ExternalTransferDirectionDeposit    ExternalTransferDirection = "deposit"
	ExternalTransferDirectionWithdrawal ExternalTransferDirection = "withdrawal"

	ExternalTransferStatusPosted   ExternalTransferStatus = "posted"
	ExternalTransferStatusPending  ExternalTransferStatus = "pending"
	ExternalTransferStatusFailed   ExternalTransferStatus = "failed"
	ExternalTransferStatusCanceled ExternalTransferStatus = "canceled"

	NormalBalanceCredit NormalBalance = "credit"
	NormalBalanceDebit  NormalBalance = "debit"

	LedgerableTypeExternalAccount LedgerableType = "external_account"
	LedgerableTypeInternalAccount LedgerableType = "internal_account"

	EntryDirectionCredit EntryDirection = "credit"
	EntryDirectionDebit  EntryDirection = "debit"
)

type TransferCommand struct {
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	IdempotencyKey IdempotencyKey
}

type PostTransferCommand struct {
	IdempotencyKey IdempotencyKey
	Amount         Amount
	ToAccountID    AccountID
	FromAccountID  AccountID
}

type PostExternalTransferCommand struct {
	IdempotencyKey            IdempotencyKey
	TransferAmount            Amount
	UserAccountID             AccountID
	SettlementAccountID       AccountID
	Rail                      PaymentRail
	ExternalReference         ExternalReference
	ExternalTransferDirection ExternalTransferDirection
}

type ReversalCommand struct {
	TransactionID  TransactionID
	IdempotencyKey IdempotencyKey
	Reason         Reason
}

type Transaction struct {
	ID             TransactionID
	Type           LedgerTransactionType
	IdempotencyKey IdempotencyKey
	CreatedAt      string
	FromAccountID  AccountID
	ToAccountID    AccountID
	Amount         Amount
	CurrencyCode   CurrencyCode
}

type Account struct {
	ID             AccountID
	Name           string
	Description    string
	CurrencyCode   CurrencyCode
	NormalBalance  NormalBalance
	LedgerableType LedgerableType
	LockVersion    int64
	Balance        int64
	CreatedAt      time.Time
}

type Entry struct {
	ID             int64
	TransactionID  TransactionID
	AccountID      AccountID
	Amount         Amount
	CreatedAt      string
	EntryDirection EntryDirection
}

type LedgerEntryInput struct {
	AccountID AccountID
	Amount    Amount
	Direction EntryDirection
}

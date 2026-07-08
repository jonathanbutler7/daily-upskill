package ledgerstore

// Named types for values that cross ledgerstore helper boundaries.
type AccountID int64
type TransactionID int64
type Amount int64
type CurrencyCode string
type IdempotencyKey string
type PaymentRail string
type ExternalReference string
type ExternalTransferDirection string
type ExternalTransferStatus string

type TransferCommand struct {
	FromAccountID  int64
	ToAccountID    int64
	Amount         int64
	IdempotencyKey string
}

type DepositFundsCommand struct {
	ToAccountID         int64
	TransferAmount      int64
	Rail                string
	ExternalReferenceID string
	IdempotencyKey      string
}

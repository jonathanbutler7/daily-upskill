package main

import (
	"context"
	"database/sql"
	"fmt"
	cmd "ledger-db/cmd"
	"ledger-db/internal/ledgerstore"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("pgx", "postgresql://ledger_db:password@localhost:5432/ledger_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tID, err := cmd.PostTransfer(ctx, db,
		ledgerstore.TransferCommand{
			FromAccountID:  2,
			ToAccountID:    3,
			Amount:         11111111,
			IdempotencyKey: "a-key-1123",
		})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("post transfer id", tID)

	transactionID, err := cmd.PostExternalTransfer(ctx, db,
		ledgerstore.PostExternalTransferCommand{
			ToAccountID:       1,
			TransferAmount:    1000,
			Rail:              "ach",
			ExternalReference: "external-id",
			IdempotencyKey:    "a-key-113",
		})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("add balance id", transactionID)
}

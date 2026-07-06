package main

import (
	"context"
	"database/sql"
	"fmt"
	cmd "ledger-db/cmd"
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
	tID, err := cmd.PostTransfer(ctx, db, cmd.TransferCommand{
		FromAccountID:  1,
		ToAccountID:    2,
		Amount:         10000,
		IdempotencyKey: "a-key-112",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("post transfer id", tID)
	tID, err = cmd.DepositFunds(ctx, db, cmd.DepositFundsCommand{
		ToAccountID:         1,
		TransferAmount:      100000,
		Rail:                "ach",
		ExternalReferenceID: "external-id",
		IdempotencyKey:      "a-key-113",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("add balance id", tID)
}

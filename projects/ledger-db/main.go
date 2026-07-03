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
	tID, err := cmd.PostTransfer(ctx, db, 1, 2, 100000, "a-key-112")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tID)
}

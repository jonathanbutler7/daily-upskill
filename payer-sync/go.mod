module weavelab.xyz/payer-sync

go 1.26.0

require (
	github.com/jackc/pgx/v5 v5.9.2
	github.com/joho/godotenv v1.5.1
	github.com/jonathanbutler7/payer-sync-data-seeder v0.0.0-00010101000000-000000000000
	github.com/pressly/goose/v3 v3.27.1
	github.com/sethvargo/go-retry v0.3.0
	github.com/stripe/stripe-go/v82 v82.5.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)

replace github.com/jonathanbutler7/payer-sync-data-seeder => ../payer-sync-data-seeder

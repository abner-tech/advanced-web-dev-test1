include .envrc

## run/api: run the cmd/api application

.PHONY: run/api
run/api:
	@echo 'Running Application...'
	@go run ./cmd/api/ -port=4000 -limiter-burst=5 -limiter-rps=2 -limiter-enabled=false -env=production -db-dsn=${TEST1_DB_DSN}

## db/psql: connect to the database using psql (terminal)
.PHONY: db/psql
db/psql: 
	psql ${TEST1_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY db/migrations/new:
	@echo 'creating migration fles for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${TEST1_DB_DSN} up
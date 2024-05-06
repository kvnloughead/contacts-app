include .envrc

# ============================================================
# HELPERS
# ============================================================

## help: print this help message
.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ============================================================
# DEVELOPMENT
# ============================================================

## run/web: run the cmd/web application
.PHONY: run/web
run/web:
	@go run ./cmd/web -db-dsn=${CONTACTS_DB_DSN} -port=4000

## db/psql: connect the the database using psql
.PHONY: db/psql
db/psql:
	psql ${CONTACTS_DB_DSN}

## db/migrations/create name=$1: generate new migration files
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all 'up' migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running all up migrations'
	migrate -path ./migrations -database ${CONTACTS_DB_DSN} up

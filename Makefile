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

# Requires global installation: `go install github.com/cosmtrek/ air@latest`  
# and the appropriate environmental variables. 
## run/air: run server using Air for live reloading. 
.PHONY: run/air
run/air:
	@export CONTACTS_DB_DSN=$(CONTACTS_DB_DSN) && export PORT=$(PORT) && air

## db/psql: connect the the database using psql
.PHONY: db/psql
db/psql:
	psql ${CONTACTS_DB_DSN}

## db/migrations/new name=$1: generate new migration files
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all 'up' migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running all up migrations'
	migrate -path ./migrations -database ${CONTACTS_DB_DSN} up

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

## run/dev: run the cmd/web application
.PHONY: run/dev
run/dev:
	@go run ./cmd/web -dsn=${CONTACTS_DB_DSN}

## db/psql: connect the the database using psql
.PHONY: db/psql
db/psql:
	psql ${CONTACTS_DB_DSN}


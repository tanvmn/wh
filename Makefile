# include variables from .envrc file
include .envrc

.PHONY: psql
psql:
	sudo systemctl start postgresql.service
	psql ${WH_DSN}

.PHONY: run/wh
run/wh:
	go run ./cmd/wh -dsn=${WH_DSN}
# include variables from .envrc file
include .envrc

.PHONY: psql/start
psql/start:
	sudo systemctl start postgresql.service
	psql ${WH_DSN}

.PHONY: psql/stop
psql/stop:
	sudo systemctl stop postgresql
	sudo systemctl status postgresql.service

.PHONY: psql/reset
psql/reset:
	sudo -u postgres psql -f sql/reset.sql
	psql ${WH_DSN} -f sql/init.sql
	psql ${WH_DSN}

.PHONY: run/wh
run/wh:
	go run ./cmd/wh -dsn=${WH_DSN}
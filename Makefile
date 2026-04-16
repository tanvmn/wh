# include variables from .envrc file
include .envrc

.PHONY: psql/start psql/stop psql/init psql/reset run/wh init
psql/start:
	sudo systemctl start postgresql.service

psql/stop:
	sudo systemctl stop postgresql.service
	sudo systemctl status postgresql.service

psql/init:
	sudo -u postgres psql -f sql/init.sql

psql/reset: psql/init
	psql ${WH_DSN} -f sql/reset.sql

init: psql/start psql/reset

run:
	go run ./cmd/wh -dsn=${WH_DSN}

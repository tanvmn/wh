# include variables from .envrc file
include .envrc

.PHONY: books
books:
	@open '/media/timng/Elements/data/BK/cs_ce/programming/go/alex_edwards/1_24/lets_go_1_24.pdf'
	@open '/media/timng/Elements/data/BK/cs_ce/programming/go/alex_edwards/1_24/lets_go_further_1_24.pdf'
	@open '/media/timng/Elements/data/BK/cs_ce/programming/go/in_practice/go_in_practice_2nd.pdf'

.PHONY: psql/start
psql/start:
	sudo systemctl start postgresql.service
	psql ${WH_DSN}

.PHONY: psql/stop
psql/stop:
	sudo systemctl stop postgresql
	sudo systemctl status postgresql.service

.PHONY: psql/init
psql/init:
	psql ${WH_DSN} -f sql/init.sql

.PHONY: run/wh
run/wh:
	go run ./cmd/wh -dsn=${WH_DSN} -smtp-password=${APP_PASS}

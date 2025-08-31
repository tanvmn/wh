# include variables from .envrc file
include .envrc

.PHONY: books
books:
	@open '/media/timng/Elements/data/BK/cs_ce/programming/go/alex_edwards/1_23/Lets Go Further (v1.23.0) (Alex Edwards) (Z-Library).pdf'
	@open '/media/timng/Elements/data/BK/cs_ce/programming/go/alex_edwards/1_24/Lets Go A Step-By-Step Guide to Creating Fast, Secure And Maintainable Web Applications With Go (Alex Edwards) (Z-Library).pdf'
	@open '/media/timng/Elements/data/BK/cs_ce/programming/go/in_practice/Go in Practice, Second Edition (Nathan Kozyra, Matt Butcher, Matt Farina) (Z-Library).pdf'

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

# include variables from .envrc file
include .envrc

.PHONY: run/wh
run/wh:
	go run ./cmd/wh -dsn=${WH_DSN}
package data

import (
	"database/sql"
	"log/slog"
)

type Data struct {
	Account Account
}

func NewData(db *sql.DB, lg *slog.Logger) *Data {
	return &Data{
		Account: Account{DB: db, Logger: lg},
	}
}

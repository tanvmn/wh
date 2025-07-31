package data

import (
	"database/sql"
	"log/slog"
)

func IDAcronyms() map[string]string {
	m := make(map[string]string)
	m["account"] = "ACC-"
	m["item"] = "ITM-"
	m["bin"] = "BIN-"
	m["tote"] = "TOT-"
	m["box"] = "BOX-"
	m["staff"] = "STF-"

	return m
}

type Data struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewData(db *sql.DB, lg *slog.Logger) *Data {
	return &Data{
		db:     db,
		logger: lg,
	}
}

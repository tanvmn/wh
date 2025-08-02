package data

import (
	"database/sql"
	"errors"
	"log/slog"
)

const (
	AccountIDCode = "ACC-"
	ItemIDCode    = "ITM-"
	BinIDCode     = "BIN-"
	ToteIDCode    = "TOT-"
	BoxIDCode     = "BOX-"
	StaffIDCode   = "STF-"
)

var (
	ErrInvalidID = errors.New("data: invalid ID")
)

func IDCodes() map[string]string {
	m := make(map[string]string)
	m["account"] = AccountIDCode 
	m["item"] = ItemIDCode
	m["bin"] = BinIDCode
	m["tote"] = ToteIDCode
	m["box"] = BoxIDCode
	m["staff"] = StaffIDCode

	return m
}

type Data struct {
	DB     *sql.DB
	logger *slog.Logger
}

func NewData(db *sql.DB, lg *slog.Logger) *Data {
	return &Data{
		DB:     db,
		logger: lg,
	}
}

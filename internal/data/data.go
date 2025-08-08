package data

import (
	"database/sql"
	"errors"
	"log/slog"
	"strconv"
)

const (
	AccountIDCode   = "ACC-"
	ItemIDCode      = "ITE-"
	BinIDCode       = "BIN-"
	ToteIDCode      = "TOT-"
	BoxIDCode       = "BOX-"
	StaffIDCode     = "STA-"
	PurchaseIDCode  = "PUR-"
	ReceiveIDCode   = "REC-"
	ResupplyIDCode  = "RES-"
	ExportIDCode    = "EXP-"
	WarehouseIDCode = "WAR-"
	StoreIDCode     = "STO-"
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

// Range returns the string of sql range syntax in a where clause
func Range(n int64) string {
	var r string

	for i := range n {
		if r == "" {
			r += "$" + strconv.FormatInt(i+1, 10)
		} else {
			r += ",$" + strconv.FormatInt(i+1, 10)
		}
	}
	r = "(" + r + ")"

	return r
}

package data

import (
	"database/sql"
	"errors"
	"log/slog"
)

var (
	ErrNoRecords = errors.New("không tìm thấy dũ liệu")
)

type Data struct {
	// Account *accountDB
	db     *sql.DB
	logger *slog.Logger
}

func NewData(db *sql.DB, lg *slog.Logger) *Data {
	return &Data{
		// Account: &accountDB{db: db, logger: lg},
		db:     db,
		logger: lg,
	}
}

package data

import (
	"database/sql"
	"log/slog"

	"github.com/tanNguyen2220022/wh/internal/model"
)

type Account struct {
	DB     *sql.DB
	Logger *slog.Logger
}

func (a *Account) Get(id int64) (*model.Account, error) {
	if id < 1 {
		return nil, sql.ErrNoRows
	}

	stmt := `select
	id,
	bdate,
	name,
	phone
	from account
	where id=$1`

	var ac model.Account
	err := a.DB.QueryRow(stmt, id).Scan(
		&ac.ID,
		&ac.BDate,
		&ac.Name,
		&ac.Phone,
	)
	if err != nil {
		a.Logger.Error(err.Error())
		return nil, err
	}

	return &ac, nil
}

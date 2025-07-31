package data

import (
	"database/sql"
	"errors"
)

type Account struct {
	ID           string  `json:"id,omitempty,omitzero"`
	BDate        string `json:"bdate,omitempty,omitzero"`
	Name         string `json:"name,omitempty,omitzero"`
	Phone        string `json:"phone,omitempty,omitzero"`
	PasswordHash []byte `json:"-"`
	WarehouseID  int64  `json:"warehouseID,omitempty,omitzero"`
	StoreID      int64  `json:"storeID,omitempty,omitzero"`
}

var (
	ErrNoAccounts = errors.New("account not found")
)

// Account only receives the integer part of the ID, not the whole string ID
func (d *Data) Account(id int64) (*Account, error) {
	stmt := `select
	'ACC-'||id,
	bdate,
	name,
	phone
	from account
	where id=$1`

	var ac Account
	err := d.db.QueryRow(stmt, id).Scan(
		&ac.ID,
		&ac.BDate,
		&ac.Name,
		&ac.Phone,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoAccounts
	} else if err != nil {
		return nil, err
	}

	return &ac, nil
}

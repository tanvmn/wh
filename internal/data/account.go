package data

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID           string `json:"id,omitempty,omitzero"`
	BDate        string `json:"bdate,omitempty,omitzero"`
	Name         string `json:"name,omitempty,omitzero"`
	Role         string `json:"role,omitempty,omitzero"`
	Phone        string `json:"phone,omitempty,omitzero"`
	PasswordHash []byte `json:"-"`
	WarehouseID  int64  `json:"warehouseID,omitempty,omitzero"`
	StoreID      int64  `json:"storeID,omitempty,omitzero"`
}

var (
	ErrNoAccounts         = errors.New("account not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func (d *Data) Account(id int64) (*Account, error) {
	stmt := `select
	'ACC-'||id,
	substring(to_char(bdate, 'YYYY-MM-DD') from 1 for 10),
	name,
	role,
	phone
	from account
	where id=$1`

	var ac Account
	err := d.DB.QueryRow(stmt, id).Scan(
		&ac.ID,
		&ac.BDate,
		&ac.Name,
		&ac.Role,
		&ac.Phone,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoAccounts
	}

	return &ac, err
}

func (d *Data) Authenticate(phone, password string) (id int64, err error) {
	var (
		passwordHash []byte
	)

	stmt := `select id, password_hash from account where phone=$1`
	err = d.DB.QueryRow(stmt, phone).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return 0, ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

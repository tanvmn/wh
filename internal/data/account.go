package data

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID           string    `json:"id,omitempty,omitzero"`
	BDate        string    `json:"bdate,omitempty,omitzero"`
	Name         string    `json:"name,omitempty,omitzero"`
	Role         string    `json:"role,omitempty,omitzero"`
	Phone        string    `json:"phone,omitempty,omitzero"`
	PasswordHash []byte    `json:"-"`
	Warehouse    Warehouse `json:"warehouse,omitempty,omitzero"`
	Store        Store     `json:"store,omitempty,omitzero"`
}

var (
	ErrNoAccounts         = errors.New("data: account not found")
	ErrInvalidCredentials = errors.New("data: invalid credentials")
)

func (d *Data) Account(id int64) (*Account, error) {
	if id < 1 {
		return nil, ErrNoAccounts
	}

	stmt := `select
	'ACC-'||id,
	substring(to_char(bdate, 'YYYY-MM-DD') from 1 for 10),
	name,
	role,
	phone,
	'WAR'||warehouse_id,
	'STO'||store_id
	from account
	where id=$1`

	var (
		ac                   Account
		warehouseID, storeID sql.NullInt64
	)
	err := d.DB.QueryRow(stmt, id).Scan(
		&ac.ID,
		&ac.BDate,
		&ac.Name,
		&ac.Role,
		&ac.Phone,
		&warehouseID,
		&storeID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoAccounts
	} else if err != nil {
		return nil, err
	}

	if warehouseID.Valid {
		ac.Warehouse.ID = WarehouseIDCode + fmt.Sprint(warehouseID.Int64)
	}
	if storeID.Valid {
		ac.Store.ID = StoreIDCode + fmt.Sprint(storeID.Int64)
	}

	return &ac, nil
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

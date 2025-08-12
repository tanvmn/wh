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

const (
	Admin          = "Admin"
	HeadAccountant = "Kế toán trưởng"
	Accountant     = "Kế toán"
	Manager        = "Thủ kho"
	Employee       = "Nhân viên"
)

var (
	ErrNoAccounts         = errors.New("data: account not found")
	ErrInvalidCredentials = errors.New("data: invalid credentials")
)

var accountSelectStmt = fmt.Sprintf(`
	select
	'%v'||id,
	substring(to_char(bdate, 'YYYY-MM-DD') from 1 for 10),
	name,
	role,
	phone,
	warehouse_id,
	store_id
	from account
	where id = $1`, AccountIDCode)

func (db *Data) Account(id string) (*Account, error) {
	i, err := id64(id, AccountIDCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoAccounts, id)
	}
	if i < 1 {
		return nil, fmt.Errorf("%w: %v", ErrNoAccounts, id)
	}

	stmt := accountSelectStmt

	var (
		ac                   Account
		warehouseID, storeID sql.NullInt64
	)
	err = db.DB.QueryRow(stmt, i).Scan(
		&ac.ID,
		&ac.BDate,
		&ac.Name,
		&ac.Role,
		&ac.Phone,
		&warehouseID,
		&storeID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %v", ErrNoAccounts, id)
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

func (db *Data) Authenticate(phone, password string) (id string, err error) {
	var (
		passwordHash []byte
	)

	stmt := `select
	'` + AccountIDCode + `'||id,
	password_hash
	from account
	where phone = $1`
	err = db.DB.QueryRow(stmt, phone).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidCredentials
	} else if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", ErrInvalidCredentials
	} else if err != nil {
		return "", err
	}

	return id, nil
}

func (db *Data) IsAccountFromWarehouse(accountID, warehouseID string) (bool, error) {
	aI, err := id64(accountID, AccountIDCode)
	if err != nil {
		return false, err
	}
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return false, err
	}

	var from bool
	stmt := `select
	true
	from account
	where
	id = $1
	and warehouse_id = $2`

	err = db.DB.QueryRow(stmt, aI, wI).Scan(&from)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

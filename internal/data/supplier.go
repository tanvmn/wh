package data

import (
	"database/sql"
	"errors"
	"fmt"
)

type Supplier struct {
	ID      string `json:"id,omitempty,omitzero"`
	Name    string `json:"name,omitempty,omitzero"`
	Address string `json:"address,omitempty,omitzero"`
	Phone   string `json:"phone,omitempty,omitzero"`
	Email   string `json:"email,omitempty,omitzero"`
}

var (
	ErrNoSuppliers = errors.New("no suppliers found")
)

func (db *Data) Supplier(id string) (*Supplier, error) {
	i, err := id64(id, SupplierIDCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoSuppliers, id)
	}
	if i < 1 {
		return nil, fmt.Errorf("%w: %v", ErrNoSuppliers, id)
	}

	stmt := fmt.Sprintf(
		`select
	'%v'||id
	,name
	,address
	,phone
	,email
	from supplier
	where
	id=$1`, SupplierIDCode)
	var sp Supplier

	err = db.DB.QueryRow(stmt, i).Scan(
		&sp.ID,
		&sp.Name,
		&sp.Address,
		&sp.Phone,
		&sp.Email,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %v", ErrNoSuppliers, id)
	} else if err != nil {
		return nil, err
	}

	return &sp, nil
}

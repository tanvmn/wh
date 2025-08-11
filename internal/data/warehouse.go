package data

import (
	"database/sql"
	"errors"
	"fmt"
)

type Warehouse struct {
	ID      string `json:"id,omitempty,omitzero"`
	Name    string `json:"name,omitempty,omitzero"`
	Address string `json:"address,omitempty,omitzero"`
	Phone   string `json:"phone,omitempty,omitzero"`
	Email   string `json:"email,omitempty,omitzero"`
}

type Bin struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
	Capacity  float32   `json:"capacity,omitempty,omitzero"`
	Shelf     int64     `json:"shelf,omitempty,omitzero"`
	Row       int64     `json:"row,omitempty,omitzero"`
	Col       int64     `json:"col,omitempty,omitzero"`
}

type Tote struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
	Capacity  float32   `json:"capacity,omitempty,omitzero"`
}

type Store struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Name      string    `json:"name,omitempty,omitzero"`
	Address   string    `json:"address,omitempty,omitzero"`
	Phone     string    `json:"phone,omitempty,omitzero"`
	Email     string    `json:"email,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
}

var (
	ErrNoWarehouses = errors.New("data: no warehouses found")
)

func (db *Data) Warehouse(id string) (*Warehouse, error) {
	i, err := id64(id, WarehouseIDCode)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrNoWarehouses, id)
	}
	if i < 1 {
		return nil, fmt.Errorf("%w, %v", ErrNoWarehouses, id)
	}

	stmt := fmt.Sprintf(
		`select
	'%v'||id
	,name
	,address
	,phone
	,email
	from warehouse
	where
	id=$1`, WarehouseIDCode)
	var wh Warehouse

	err = db.DB.QueryRow(stmt, i).Scan(
		&wh.ID,
		&wh.Name,
		&wh.Address,
		&wh.Phone,
		&wh.Email,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w, %v", ErrNoWarehouses, id)
	} else if err != nil {
		return nil, err
	}

	return &wh, nil
}

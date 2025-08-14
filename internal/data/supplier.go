package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var sp Supplier

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
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

func (db *Data) Suppliers() ([]Supplier, error) {
	stmt := fmt.Sprintf(`
	select
	'%v'||id
	,name
	,address
	,phone
	,email
	from supplier
	`, SupplierIDCode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var ss []Supplier

	rows, err := db.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err)
		}
	}()

	for rows.Next() {
		var s Supplier

		err = rows.Scan(
			&s.ID,
			&s.Name,
			&s.Address,
			&s.Phone,
			&s.Email,
		)
		if err != nil {
			return nil, err
		}

		ss = append(ss, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ss, nil
}

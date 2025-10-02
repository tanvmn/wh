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
	Items   []Item `json:"items,omitempty,omitzero"`
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
		db.logger.Error(err.Error())
		return nil, fmt.Errorf("%w: %v", ErrNoSuppliers, id)
	} else if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	sp.Items, err = db.SupplierItems(sp.ID)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	return &sp, nil
}

func (db *Data) SupplierItems(supplierID string) ([]Item, error) {
	sI, err := id64(supplierID, SupplierIDCode)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	stmt := `
	select
	gtin
	from supplier_item
	where supplier_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, sI)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}
	defer func() {
		if err2 := rows.Close(); err2 != nil {
			panic(err2)
		}
	}()

	var is []Item

	for rows.Next() {
		i := new(Item)

		err = rows.Scan(
			&i.GTIN,
		)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}

		i, err = db.Item(i.GTIN)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}

		is = append(is, *i)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return is, nil
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
			panic(err2)
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

		// Add the items that supplier supplies
		gtins, err := db.GTINsBySupplier(s.ID)
		if err != nil {
			return nil, err
		}
		for _, gtin := range gtins {
			i, err := db.Item(gtin)
			if err != nil {
				return nil, err
			}
			s.Items = append(s.Items, *i)
		}

		ss = append(ss, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ss, nil
}

func addSupplier(tx *sql.Tx, input *Supplier) (id string, err error) {
	stmt := `
	insert into supplier (name, address, phone, email) values
	($1, $2, $3, $4)
	returning $5||id
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tx.QueryRowContext(ctx, stmt, input.Name, input.Address, input.Phone, input.Email, SupplierIDCode).Scan(
		&id,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func addSupplierItems(tx *sql.Tx, input *Supplier) error {
	sI, err := id64(input.ID, SupplierIDCode)
	if err != nil {
		return err
	}

	stmt := `
	insert into supplier_item (supplier_id, gtin) values
	($1, $2)
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, i := range input.Items {
		_, err := tx.ExecContext(ctx, stmt, sI, i.GTIN)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Data) AddSupplier(input *Supplier) (id string, err error) {
	tx, err := db.DB.Begin()
	if err != nil {
		db.logger.Error(err.Error())
		return "", err
	}
	defer tx.Rollback()

	input.ID, err = addSupplier(tx, input)
	if err != nil {
		db.logger.Error(err.Error())
		return "", err
	}

	err = addSupplierItems(tx, input)
	if err != nil {
		db.logger.Error(err.Error())
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return input.ID, nil
}

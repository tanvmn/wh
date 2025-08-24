package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Receive struct {
	ID         string         `json:"id,omitempty,omitzero"`
	Purchase   Purchase       `json:"purchase,omitempty,omitzero"`
	Account    Account        `json:"account,omitempty,omitzero"`
	ExpectedAt string         `json:"expectedAt,omitempty,omitzero"`
	ActualAt   string         `json:"actualAt,omitempty,omitzero"`
	CreatedAt  string         `json:"createdAt,omitempty,omitzero"`
	Transfer   Transfer       `json:"transfer,omitempty,omitzero"`
	Version    int            `json:"version,omitempty,omitzero"`
	Items      []ItemQuantity `json:"items,omitempty,omitzero"`
}

var (
	ErrNoReceives     = errors.New("data: no receives found")
	ErrNoReceiveItems = errors.New("data: no receive items found")
)

func (db *Data) ReceiveItemsByPurchase(purchaseID string) ([]ItemQuantity, error) {
	i, err := id64(purchaseID, PurchaseIDCode)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	select
	ri.gtin
	,characteristic
	,volume
	,weight
	,brand
	,material
	,color
	,size
	,price
	,currency
	,type
	,shelf_life
	,type||', '||brand||', màu '||color||', cỡ '||size||', '||material
	,sum(quantity)
	from receive_item as ri
	join item on item.gtin = ri.gtin
	where purchase_id = $1
	group by
	ri.gtin
	,characteristic
	,volume
	,weight
	,brand
	,material
	,color
	,size
	,price
	,currency
	,type
	,shelf_life
	;
	`

	rows, err := db.DB.QueryContext(ctx, stmt, i)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err2)
		}
	}()

	var is []ItemQuantity
	for rows.Next() {
		var iq ItemQuantity

		err = rows.Scan(
			&iq.Item.GTIN,
			&iq.Item.Characteristic,
			&iq.Item.Volume,
			&iq.Item.Weight,
			&iq.Item.Brand,
			&iq.Item.Material,
			&iq.Item.Color,
			&iq.Item.Size,
			&iq.Item.Price,
			&iq.Item.Currency,
			&iq.Item.Type,
			&iq.Item.ShelfLife,
			&iq.Item.Name,
			&iq.Quantity,
		)
		is = append(is, iq)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return is, nil
}

func delReceiveItemsByPurchase(tx *sql.Tx, purchaseID64 int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	delete from receive_item
	where
	purchase_id = $1`

	_, err := tx.ExecContext(ctx, stmt, purchaseID64)
	if err != nil {
		return err
	}

	return nil
}

func delReceivesByPurchase(tx *sql.Tx, purchaseID64 int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	delete from receive
	where
	purchase_id = $1`

	_, err := tx.ExecContext(ctx, stmt, purchaseID64)
	if err != nil {
		return err
	}

	return nil
}

func addReceiveItems(tx *sql.Tx, rc *Receive) error {
	pI, err := id64(rc.Purchase.ID, PurchaseIDCode)
	if err != nil {
		return err
	}
	rI, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	insert into receive_item (purchase_id, gtin, receive_id, quantity) values
	($1,$2,$3,$4)
	;
	`
	for _, it := range rc.Items {
		_, err = tx.ExecContext(ctx, stmt, pI, it.Item.GTIN, rI, it.Quantity)
		if err != nil {
			return err
		}
	}

	return nil
}

func addReceive(tx *sql.Tx, rc *Receive) error {
	aI, err := id64(rc.Account.ID, AccountIDCode)
	if err != nil {
		return err
	}
	pI, err := id64(rc.Purchase.ID, PurchaseIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	insert into receive (purchase_id, account_id, expected_at) values
	($1,$2,$3)
	returning '%v'||id, version
	;`, ReceiveIDCode)

	err = tx.QueryRowContext(ctx, stmt, pI, aI, rc.ExpectedAt).Scan(&rc.ID, &rc.Version)
	if err != nil {
		return err
	}

	err = addReceiveItems(tx, rc)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) AddReceive(rc *Receive) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = addReceive(tx, rc)
	if err != nil {
		return err
	}

	err = db.UnclaimReceiveAddOwner(rc.Purchase.ID, rc.Account.ID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

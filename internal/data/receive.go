package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Receive struct {
	ID               string         `json:"id,omitempty,omitzero"`
	ExpectedAt       string         `json:"expectedAt,omitempty,omitzero"`
	ActualAt         string         `json:"actualAt,omitempty,omitzero"`
	CreatedAt        string         `json:"createdAt,omitempty,omitzero"`
	PutawayAt        string         `json:"putawayAt,omitempty,omitzero"`
	VoucherID        string         `json:"voucherID,omitempty,omitzero"`
	Version          int            `json:"version,omitempty,omitzero"`
	Note             string         `json:"note,omitempty,omitzero"`
	ExpectedQuantity int64          `json:"expectedQuantity,omitempty,omitzero"`
	ReceivedQuantity int64          `json:"receivedQuantity,omitempty,omitzero"`
	Items            []ItemQuantity `json:"items,omitempty,omitzero"`
	Transfer         `json:"transfer,omitempty,omitzero"`
	Purchase         `json:"purchase,omitempty,omitzero"`
	Account          `json:"account,omitempty,omitzero"`
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
	;`

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
		if err != nil {
			return nil, err
		}
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

func setReceiveVoucherID(tx *sql.Tx, receiveID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	update receive set voucher_id='VOU-00'||$1 where id=$1
	;`

	_, err := tx.ExecContext(ctx, stmt, receiveID)
	if err != nil {
		return err
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
	insert into receive (purchase_id, account_id, expected_at, voucher_id) values
	($1,$2,$3,$4)
	returning '%v'||id, version
	;`, ReceiveIDCode)

	err = tx.QueryRowContext(ctx, stmt, pI, aI, rc.ExpectedAt, rc.VoucherID).Scan(&rc.ID, &rc.Version)
	if err != nil {
		return err
	}

	// set receive's voucher_id with format VOU-<receive_id>
	rI, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
	}
	err = setReceiveVoucherID(tx, rI)
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

func (db *Data) ReceiveItems(rc *Receive) error {
	i, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
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
	,img_fspath
	,quantity
	from receive_item as ri
	join item on item.gtin = ri.gtin
	join receive on receive.id = ri.receive_id
	where receive_id = $1
	;`

	rows, err := db.DB.QueryContext(ctx, stmt, i)
	if err != nil {
		return err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err2)
		}
	}()

	rc.Items = nil
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
			&iq.ImgFSPath,
			&iq.Quantity,
		)
		if err != nil {
			return err
		}
		rc.Items = append(rc.Items, iq)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// db.Serial

	return nil
}

func (db *Data) Receive(id string) (*Receive, error) {
	i, err := id64(id, ReceiveIDCode)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	select
	'%v'||receive.id
	,'%v'||purchase_id
	,'%v'||account_id
	,expected_at
	,actual_at
	,created_at
	,'%v'||transfer_id
	,receive.version
	,note
	,voucher_id
	from receive
	where receive.id = $1
	;`,
		ReceiveIDCode,
		PurchaseIDCode,
		AccountIDCode,
		TransferIDCode,
	)
	var (
		rc         Receive
		transferID sql.NullString
	)

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
		&rc.ID,
		&rc.Purchase.ID,
		&rc.Account.ID,
		&rc.ExpectedAt,
		&rc.ActualAt,
		&rc.CreatedAt,
		&transferID,
		&rc.Version,
		&rc.Note,
		&rc.VoucherID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", ErrNoReceives, id)
		}
		return nil, err
	}

	rc.ExpectedAt = rc.ExpectedAt[:16]
	rc.CreatedAt = rc.CreatedAt[:16]

	err = db.ReceiveItems(&rc)
	if err != nil {
		return nil, err
	}

	pc, err := db.Purchase(rc.Purchase.ID)
	if err != nil {
		return nil, err
	}
	rc.Purchase = *pc

	ac, err := db.Account(rc.Account.ID)
	if err != nil {
		return nil, err
	}
	rc.Account = *ac

	if transferID.Valid {
		tf, err := db.Transfer(rc.Transfer.ID)
		if err != nil {
			if !errors.Is(err, ErrNoTransfers) {
				return nil, err
			}
		}
		rc.Transfer = *tf
	}

	err = db.MaxReceiveQuantity(&rc)
	if err != nil {
		return nil, err
	}

	err = db.SerialsByReceive(&rc)
	if err != nil {
		return nil, err
	}

	db.UpdateReceiveQuantity(&rc)

	return &rc, nil
}

func (db *Data) ReceivesByPurchase(purchaseID string) ([]Receive, error) {
	i, err := id64(purchaseID, PurchaseIDCode)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	select
	'%v'||receive.id
	,'%v'||purchase_id
	,'%v'||account_id
	,to_char(receive.expected_at, 'DD-MM-YYYY HH24:MI')
	,to_char(receive.actual_at, 'DD-MM-YYYY HH24:MI')
	,to_char(receive.created_at, 'DD-MM-YYYY HH24:MI')
	,'%v'||transfer_id
	,receive.version
	,note
	from receive
	join account on account.id = receive.account_id
	where receive.purchase_id = $1
	;`,
		ReceiveIDCode,
		PurchaseIDCode,
		AccountIDCode,
		TransferIDCode,
	)
	var (
		rs         []Receive
		transferID sql.NullString
	)

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

	for rows.Next() {
		var rc Receive
		err = rows.Scan(
			&rc.ID,
			&rc.Purchase.ID,
			&rc.Account.ID,
			&rc.ExpectedAt,
			&rc.ActualAt,
			&rc.CreatedAt,
			&transferID,
			&rc.Version,
			&rc.Note,
		)
		if err != nil {
			return nil, err
		}

		if transferID.Valid {
			rc.Transfer.ID = transferID.String
			tf, err := db.Transfer(rc.Transfer.ID)
			if err != nil {
				if !errors.Is(err, ErrNoTransfers) {
					return nil, err
				}
			}
			rc.Transfer = *tf
		}

		pc, err := db.Purchase(rc.Purchase.ID)
		if err != nil {
			return nil, err
		}
		rc.Purchase = *pc

		ac, err := db.Account(rc.Account.ID)
		if err != nil {
			return nil, err
		}
		rc.Account = *ac

		err = db.ReceiveItems(&rc)
		if err != nil {
			return nil, err
		}

		err = db.MaxReceiveQuantity(&rc)
		if err != nil {
			return nil, err
		}

		err = db.SerialsByReceive(&rc)
		if err != nil {
			return nil, err
		}

		db.UpdateReceiveQuantity(&rc)

		rs = append(rs, rc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rs, nil
}

func (db *Data) MaxReceiveQuantity(rc *Receive) error {
	pis, err := db.UnreceivedPurchaseItems(&rc.Purchase)
	if err != nil {
		return err
	}

	for i := range rc.Items {
		for _, pi := range pis {
			if pi.Item.GTIN == rc.Items[i].Item.GTIN {
				rc.Items[i].MaxReceiveQuantity = rc.Items[i].Quantity + pi.Quantity
				break
			}
		}
		if rc.Items[i].MaxReceiveQuantity == 0 {
			rc.Items[i].MaxReceiveQuantity = rc.Items[i].Quantity
		}
	}

	return nil
}

func delReceiveItems(tx *sql.Tx, id64 int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	delete from receive_item
	where
	receive_id = $1
	;`

	_, err := tx.ExecContext(ctx, stmt, id64)
	if err != nil {
		return err
	}

	return nil
}

func setReceiveItems(tx *sql.Tx, rc *Receive) error {
	i, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
	}

	err = delReceiveItems(tx, i)
	if err != nil {
		return err
	}

	err = addReceiveItems(tx, rc)
	if err != nil {
		return err
	}

	return nil
}

func setReceive(tx *sql.Tx, rc *Receive) error {
	i, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	update receive set
	expected_at = $1
	, version = version + 1
	, voucher_id = $2
	where
	id = $3
	and version = $4
	returning version
	;`
	var version int

	err = tx.QueryRowContext(ctx, stmt, rc.ExpectedAt, rc.VoucherID, i, rc.Version).Scan(&version)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) SetReceive(rc *Receive) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = setReceive(tx, rc)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoReceives
		}
		return err
	}

	err = setReceiveItems(tx, rc)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func delReceive(tx *sql.Tx, id64 int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	delete from receive
	where
	id = $1`

	_, err := tx.ExecContext(ctx, stmt, id64)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) DelReceive(id string) error {
	i, err := id64(id, ReceiveIDCode)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = delReceiveItems(tx, i)
	if err != nil {
		return err
	}

	err = delReceive(tx, i)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) Receives(warehouseID string) ([]Receive, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	select
	'%v'||receive.id
	,'%v'||purchase_id
	,'%v'||receive.account_id
	,to_char(receive.expected_at, 'DD-MM-YYYY HH24:MI')
	,to_char(receive.actual_at, 'DD-MM-YYYY HH24:MI')
	,to_char(receive.created_at, 'DD-MM-YYYY HH24:MI')
	,'%v'||transfer_id
	,receive.version
	,receive.note
	from receive
	join account on account.id = receive.account_id
	join purchase on purchase.id = receive.purchase_id
	where purchase.warehouse_id = $1
	;`,
		ReceiveIDCode,
		PurchaseIDCode,
		AccountIDCode,
		TransferIDCode,
	)
	var (
		rs         []Receive
		transferID sql.NullString
	)

	rows, err := db.DB.QueryContext(ctx, stmt, wI)
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
		var rc Receive
		err = rows.Scan(
			&rc.ID,
			&rc.Purchase.ID,
			&rc.Account.ID,
			&rc.ExpectedAt,
			&rc.ActualAt,
			&rc.CreatedAt,
			&transferID,
			&rc.Version,
			&rc.Note,
		)
		if err != nil {
			return nil, err
		}

		if transferID.Valid {
			rc.Transfer.ID = transferID.String
			tf, err := db.Transfer(rc.Transfer.ID)
			if err != nil {
				if !errors.Is(err, ErrNoTransfers) {
					return nil, err
				}
			}
			rc.Transfer = *tf
		}

		pc, err := db.Purchase(rc.Purchase.ID)
		if err != nil {
			return nil, err
		}
		rc.Purchase = *pc

		ac, err := db.Account(rc.Account.ID)
		if err != nil {
			return nil, err
		}
		rc.Account = *ac

		err = db.ReceiveItems(&rc)
		if err != nil {
			return nil, err
		}

		err = db.MaxReceiveQuantity(&rc)
		if err != nil {
			return nil, err
		}

		err = db.SerialsByReceive(&rc)
		if err != nil {
			return nil, err
		}

		db.UpdateReceiveQuantity(&rc)

		rs = append(rs, rc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rs, nil
}

func setActualAt(tx *sql.Tx, rc *Receive) error {
	rI, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	update receive
	set
	actual_at = localtimestamp
	where id = $1
	;`

	_, err = tx.ExecContext(ctx, stmt, rI)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) SetActualAt(rc *Receive) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = setActualAt(tx, rc)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func setReceiveItemsNote(tx *sql.Tx, rc *Receive) error {
	rI, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
	}
	pI, err := id64(rc.Purchase.ID, PurchaseIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	update receive_item
	set
	note = $1
	where receive_id = $2
	and purchase_id = $3
	and gtin = $4
	;`

	for _, i := range rc.Items {
		if i.Note != "none" {
			_, err = tx.ExecContext(ctx, stmt, i.Note, rI, pI, i.Item.GTIN)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *Data) SetReceiveItemsNote(rc *Receive) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = setReceiveItemsNote(tx, rc)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) UnprocessedReceivesByPurchase(purchaseID string) ([]Receive, error) {
	temp, err := db.ReceivesByPurchase(purchaseID)
	if err != nil {
		return nil, err
	}

	var rs []Receive
	for _, r := range temp {
		if r.ActualAt[:16] == "1000-01-01 00:00" {
			rs = append(rs, r)
		}
	}

	return rs, nil
}

func (db *Data) UpdateReceiveQuantity(rc *Receive) {
	for _, iq := range rc.Items {
		rc.ExpectedQuantity += iq.Quantity
		rc.ReceivedQuantity += int64(len(iq.Serials))
	}
}

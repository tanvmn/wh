package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNoPurchases      = errors.New("data: no purchases found")
	ErrNoPurchaseItems  = errors.New("data: no purchase items found")
	ErrPurchaseReceived = errors.New("data: purchase received")
)

type Purchase struct {
	ID         string         `json:"id,omitempty,omitzero"`
	ExpectedAt string         `json:"expectedAt,omitempty,omitzero"`
	Status     string         `json:"status,omitempty,omitzero"`
	CreatedAt  string         `json:"createdAt,omitempty,omitzero"`
	Version    int            `json:"version,omitempty,omitzero"`
	Items      []ItemQuantity `json:"items,omitempty,omitzero"`
	Account    `json:"account,omitempty,omitzero"`
	Warehouse  `json:"warehouse,omitempty,omitzero"`
	Supplier   `json:"supplier,omitempty,omitzero"`
}

func addPurchase(tx *sql.Tx, pc *Purchase) (id string, version int, err error) {
	wI, err := id64(pc.Warehouse.ID, WarehouseIDCode)
	if err != nil {
		return "", 0, err
	}
	aI, err := id64(pc.Account.ID, AccountIDCode)
	if err != nil {
		return "", 0, err
	}
	sI, err := id64(pc.Supplier.ID, SupplierIDCode)
	if err != nil {
		return "", 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	insert into purchase (warehouse_id, account_id, supplier_id, expected_at) values
	($1, $2, $3, $4)
	returning '%v'||id, version`,
		PurchaseIDCode,
	)

	err = tx.QueryRowContext(ctx, stmt, wI, aI, sI, pc.ExpectedAt).Scan(&id, &version)
	if err != nil {
		return "", 0, err
	}

	pc.ID = id
	pc.Version = version

	err = addPurchaseItems(tx, pc)
	if err != nil {
		return "", 0, err
	}

	return id, version, nil
}

func addPurchaseItems(tx *sql.Tx, pc *Purchase) error {
	pI, err := id64(pc.ID, PurchaseIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	insert into purchase_item (purchase_id, gtin, quantity) values
	($1,$2,$3)`
	for _, i := range pc.Items {
		_, err := tx.ExecContext(ctx, stmt, pI, i.Item.GTIN, i.Quantity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Data) CheckCapacity(is []ItemQuantity, warehouseID string) (bool, error) {
	capacity, err := db.Capacity(warehouseID)
	if err != nil {
		return false, err
	}

	used, err := db.UsedVolume(warehouseID)
	if err != nil {
		return false, err
	}

	volume, err := db.Volume(is)
	if err != nil {
		return false, err
	}

	return (capacity - used) >= volume, nil
}

func (db *Data) AddPurchase(pc *Purchase) (id string, version int, err error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()

	id, version, err = addPurchase(tx, pc)
	if err != nil {
		return "", 0, err
	}

	err = tx.Commit()
	if err != nil {
		return "", 0, err
	}

	return id, version, nil
}

func (db *Data) Purchase(id string) (*Purchase, error) {
	i, err := id64(id, PurchaseIDCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoPurchases, err)
	}

	stmt := fmt.Sprintf(`
	select
	'%v'||id
	,'%v'||warehouse_id
	,'%v'||account_id
	,'%v'||supplier_id
	,expected_at
	,created_at
	,purchase.version
	,status
	from purchase
	where id=$1
	`,
		PurchaseIDCode,
		WarehouseIDCode,
		AccountIDCode,
		SupplierIDCode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var pc Purchase

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
		&pc.ID,
		&pc.Warehouse.ID,
		&pc.Account.ID,
		&pc.Supplier.ID,
		&pc.ExpectedAt,
		&pc.CreatedAt,
		&pc.Version,
		&pc.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w, ID %v", ErrNoPurchases, id)
	} else if err != nil {
		return nil, err
	}

	// pc.CreatedAt, err = util.FormatRFC3339(pc.CreatedAt, time.DateTime)
	// if err != nil {
	// 	return nil, err
	// }
	pc.CreatedAt = pc.CreatedAt[:16]

	// pc.ExpectedAt, err = util.FormatRFC3339(pc.ExpectedAt, time.DateTime)
	// if err != nil {
	// 	return nil, err
	// }
	pc.ExpectedAt = pc.ExpectedAt[:16]

	// Add purchase items
	err = db.PurchaseItems(&pc)
	if err != nil {
		return nil, err
	}

	// Add warehouse
	wh, err := db.Warehouse(pc.Warehouse.ID)
	if err != nil {
		return nil, err
	}
	pc.Warehouse = *wh

	// Add supplier
	sp, err := db.Supplier(pc.Supplier.ID)
	if err != nil {
		return nil, err
	}
	pc.Supplier = *sp

	// Add account
	ac, err := db.Account(pc.Account.ID)
	if err != nil {
		return nil, err
	}
	pc.Account = *ac

	return &pc, nil
}

func (db *Data) PurchaseItems(pc *Purchase) error {
	i, err := id64(pc.ID, PurchaseIDCode)
	if err != nil {
		return err
	}

	stmt := `
	select
	pi.gtin
	,characteristic
	,volume
	,weight
	,brand
	,material
	,color
	,size
	,price
	,currency
	,shelf_life
	,img_fspath
	,pi.version
	,type
	,quantity
	,type||', '||brand||', màu '||color||', cỡ '||size||', '||characteristic
	from purchase_item as pi
	join item on item.gtin = pi.gtin
	join purchase on purchase.id = pi.purchase_id
	where purchase_id = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	pc.Items = nil
	for rows.Next() {
		var i ItemQuantity

		err = rows.Scan(
			&i.Item.GTIN,
			&i.Item.Characteristic,
			&i.Item.Volume,
			&i.Item.Weight,
			&i.Item.Brand,
			&i.Item.Material,
			&i.Item.Color,
			&i.Item.Size,
			&i.Item.Price,
			&i.Item.Currency,
			&i.Item.ShelfLife,
			&i.Item.ImgFSPath,
			&i.Item.Version,
			&i.Item.Type,
			&i.Quantity,
			&i.Item.Name,
		)
		if err != nil {
			return err
		}

		pc.Items = append(pc.Items, i)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}

func (db *Data) BeforePurchaseExpectedAt(purchaseID, datetime string) (bool, string, error) {
	id, err := id64(purchaseID, PurchaseIDCode)
	if err != nil {
		return false, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `select $1 <= expected_at, expected_at from purchase where id = $2`
	var (
		before     bool
		expectedAt string
	)

	err = db.DB.QueryRowContext(ctx, stmt, datetime, id).Scan(
		&before,
		&expectedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", ErrNoPurchases
	} else if err != nil {
		return false, "", err
	}

	return before, expectedAt, nil
}

func (db *Data) PurchaseReceived(purchaseID string) (bool, error) {
	i, err := id64(purchaseID, PurchaseIDCode)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	select
	true
	from purchase
	where
	id = $1
	and (status = '%v' or status = '%v');
	`,
		AwaitingResponse,
		AwaitingReceive)
	var received bool

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(&received)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrNoPurchases
	} else if err != nil {
		return false, err
	}

	return received, nil
}

func setPurchase(tx *sql.Tx, pc *Purchase) error {
	i, err := id64(pc.ID, PurchaseIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	update purchase
	set
	expected_at = $1
	, version = version + 1
	where id = $2
	and version = $3
	returning version;
	`
	var version int

	err = tx.QueryRowContext(ctx, stmt, pc.ExpectedAt, i, pc.Version).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSetConflict
	} else if err != nil {
		return err
	}

	return nil
}

func setPurchaseItem(tx *sql.Tx, pc *Purchase) error {
	i, err := id64(pc.ID, PurchaseIDCode)
	if err != nil {
		return err
	}

	err = delPurchaseItems(tx, i)
	if err != nil {
		return err
	}

	err = addPurchaseItems(tx, pc)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) SetPurchase(pc *Purchase) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = setPurchase(tx, pc)
	if err != nil {
		return err
	}

	err = setPurchaseItem(tx, pc)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func delPurchaseItems(tx *sql.Tx, purchaseID64 int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	delete from purchase_item
	where
	purchase_id = $1`

	result, err := tx.ExecContext(ctx, stmt, purchaseID64)
	if err != nil {
		return err
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrNoPurchaseItems
	}

	return nil
}

func delPurchase(tx *sql.Tx, purchaseID64 int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	delete from purchase
	where
	id = $1`

	result, err := tx.ExecContext(ctx, stmt, purchaseID64)
	if err != nil {
		return err
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrNoPurchases
	}

	return nil
}

func (db *Data) DelPurchase(purchaseID string) error {
	i, err := id64(purchaseID, PurchaseIDCode)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = delReceiveItemsByPurchase(tx, i)
	if err != nil {
		return err
	}

	err = delReceivesByPurchase(tx, i)
	if err != nil {
		return err
	}

	err = delPurchaseItems(tx, i)
	if err != nil {
		return err
	}

	err = delPurchase(tx, i)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

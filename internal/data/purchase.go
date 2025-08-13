package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	AwaitingResponse = "Chờ phản hồi"
	AwaitingReceive  = "Chờ nhập"
	Receiving        = "Đang nhập"
	Ended            = "Kết thúc"
	Declined         = "Từ chối"
)

type Purchase struct {
	ID         string         `json:"purchase,omitempty,omitzero"`
	Warehouse  Warehouse      `json:"warehouse,omitempty,omitzero"`
	Account    Account        `json:"account,omitempty,omitzero"`
	ExpectedAt string         `json:"expectedAt,omitempty,omitzero"`
	Status     string         `json:"status,omitempty,omitzero"`
	Supplier   Supplier       `json:"supplier,omitempty,omitzero"`
	CreatedAt  string         `json:"createdAt,omitempty,omitzero"`
	Items      []ItemQuantity `json:"items,omitempty,omitzero"`
	Version    int            `json:"version,omitempty,omitzero"`
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
	insert into purchase (warehouse_id, account_id, supplier_id, expected_dtime) values
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

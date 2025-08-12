package data

import (
	"database/sql"
	"fmt"
)

const (
	AwaitingResponse = "Chờ phản hồi"
	AwaitingReceive  = "Chờ nhập"
	Receiving        = "Đang nhập"
	Ended            = "Kết thúc"
	Declined         = "Từ chối"
)

type Purchase struct {
	ID         string    `json:"purchase,omitempty,omitzero"`
	Warehouse  Warehouse `json:"warehouse,omitempty,omitzero"`
	Account    Account   `json:"account,omitempty,omitzero"`
	ExpectedAt string    `json:"expectedAt,omitempty,omitzero"`
	Status     string    `json:"status,omitempty,omitzero"`
	Supplier   Supplier  `json:"supplier,omitempty,omitzero"`
	CreatedAt  string    `json:"createdAt,omitempty,omitzero"`
	Items      []struct {
		Item     Item  `json:"item,omitempty,omitzero"`
		Quantity int64 `json:"quantity,omitempty,omitzero"`
	} `json:"items,omitempty,omitzero"`
}

type Receive struct {
	ID         string   `json:"id,omitempty,omitzero"`
	Purchase   Purchase `json:"purchase,omitempty,omitzero"`
	Account    Account  `json:"account,omitempty,omitzero"`
	ExpectedAt string   `json:"expectedAt,omitempty,omitzero"`
	ActualAt   string   `json:"actualAt,omitempty,omitzero"`
	CreatedAt  string   `json:"createdAt,omitempty,omitzero"`
	Transfer   Transfer `json:"transfer,omitempty,omitzero"`
	Items      []struct {
		Item     Item  `json:"item,omitempty,omitzero"`
		Quantity int64 `json:"quantity,omitempty,omitzero"`
	} `json:"items,omitempty,omitzero"`
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

	stmt := fmt.Sprintf(`
	insert into purchase (warehouse_id, account_id, supplier_id, expected_dtime) values
	($1, $2, $3, $4)
	returning '%v'||id, version`,
		PurchaseIDCode,
	)

	err = tx.QueryRow(stmt, wI, aI, sI, pc.ExpectedAt).Scan(&id, &version)
	if err != nil {
		return "", 0, err
	}

	pc.ID = id

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

	stmt := `
	insert into purchase_item (purchase_id, gtin, quantity) values
	($1,$2,$3)`
	for _, i := range pc.Items {
		_, err := tx.Exec(stmt, pI, i.Item.GTIN, i.Quantity)
		if err != nil {
			return err
		}
	}

	return nil
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

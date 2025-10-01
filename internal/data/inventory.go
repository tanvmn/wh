package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Inventory struct {
	ID               string            `json:"id,omitempty,omitzero"`
	CreatedAt        string            `json:"createdAt,omitempty,omitzero"`
	ExpectedAt       string            `json:"expectedAt,omitempty,omitzero"`
	StartedAt        string            `json:"startedAt,omitempty,omitzero"`
	EndedAt          string            `json:"endedAt,omitempty,omitzero"`
	Note             string            `json:"note,omitempty,omitzero"`
	Balanced         bool              `json:"balanced,omitempty,omitzero"`
	Version          int               `json:"version,omitempty,omitzero"`
	Items            []ItemQuantity    `json:"items,omitempty,omitzero"`
	InventorySerials []InventorySerial `json:"serials,omitempty,omitzero"`
	Account          `json:"account,omitempty,omitzero"`
}

type InventorySerial struct {
	Note      string `json:"note,omitempty,omitzero"`
	Result    string `json:"result,omitempty,omitzero"`
	Inventory `json:"inventory,omitempty,omitzero"`
	Serial    `json:"serial,omitempty,omitzero"`
}

const (
	InventoryUnchecked = "unchecked"
	InventoryFound     = "CÓ"
	InventoryNotFound  = "KHÔNG"
)

var (
	ErrNoInventories = errors.New("no inventories found")
)

func (i *Inventory) SerialsByGTIN(gtin string) []InventorySerial {
	var iss []InventorySerial

	for _, is := range i.InventorySerials {
		if is.Serial.Item.GTIN == gtin {
			iss = append(iss, is)
		}
	}

	return iss
}

func (i *Inventory) FoundSerialsByGTIN(gtin string) []InventorySerial {
	var iss []InventorySerial

	for _, is := range i.InventorySerials {
		if is.Serial.Item.GTIN == gtin && is.Result == InventoryFound {
			iss = append(iss, is)
		}
	}

	return iss
}

func (i *Inventory) NotFoundSerialsByGTIN(gtin string) []InventorySerial {
	var iss []InventorySerial

	for _, is := range i.InventorySerials {
		if is.Serial.Item.GTIN == gtin && is.Result == InventoryNotFound {
			iss = append(iss, is)
		}
	}

	return iss
}

func addInventory(tx *sql.Tx, inventoryAddRequest *Inventory) (id string, err error) {
	wI, err := id64(inventoryAddRequest.Warehouse.ID, WarehouseIDCode)
	if err != nil {
		return "", err
	}
	aI, err := id64(inventoryAddRequest.Account.ID, AccountIDCode)
	if err != nil {
		return "", err
	}

	stmt := `
	insert into inventory (expected_at, warehouse_id, account_id) values
	($1, $2, $3)
	returning $4||id
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tx.QueryRowContext(ctx, stmt, inventoryAddRequest.ExpectedAt, wI, aI, InventoryIDCode).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func addInventorySerials(tx *sql.Tx, inventoryAddRequest *Inventory) error {
	iI, err := id64(inventoryAddRequest.ID, InventoryIDCode)
	if err != nil {
		return err
	}

	stmt := `
	insert into inventory_serial (serial, inventory_id) values 
	($1, $2)
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, is := range inventoryAddRequest.InventorySerials {
		_, err := tx.ExecContext(ctx, stmt, is.Serial.NanoID, iI)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Data) AddInventory(inventoryAddRequest *Inventory) (id string, err error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Remember to assign the newly created inventory id (string) before calling addInventorySerial
	id, err = addInventory(tx, inventoryAddRequest)
	if err != nil {
		db.logger.Error(err.Error())
		return "", err
	}
	inventoryAddRequest.ID = id

	// Add the UnexportedSerialsByGTINAndWarehouse to each inventoryAddRequest's Item
	for _, iq := range inventoryAddRequest.Items {
		ss, err := db.UnexportedSerialsByGTINAndWarehouse(inventoryAddRequest.Warehouse.ID, iq.Item.GTIN)
		if err != nil {
			db.logger.Error(err.Error())
			return "", err
		}
		for _, s := range ss {
			var is InventorySerial
			is.Serial = s
			inventoryAddRequest.InventorySerials = append(inventoryAddRequest.InventorySerials, is)
		}
	}

	err = addInventorySerials(tx, inventoryAddRequest)
	if err != nil {
		db.logger.Error(err.Error())
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		db.logger.Error(err.Error())
		return "", err
	}

	return id, nil
}

func (db *Data) InventoryItems(inventoryID string) ([]ItemQuantity, error) {
	iI, err := id64(inventoryID, InventoryIDCode)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	stmt := `
	select
	distinct serial.gtin
	from inventory_serial as isr
	join serial on serial.nanoid = isr.serial
	where inventory_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var iqs []ItemQuantity

	rows, err := db.DB.QueryContext(ctx, stmt, iI)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err2)
		}
	}()

	for rows.Next() {
		var iq ItemQuantity

		err = rows.Scan(
			&iq.Item.GTIN,
		)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}

		i, err := db.Item(iq.Item.GTIN)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}
		iq.Item = *i

		iqs = append(iqs, iq)
	}

	err = rows.Err()
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	return iqs, nil
}

func (db *Data) InventorySerials(inventoryID string) ([]InventorySerial, error) {
	iI, err := id64(inventoryID, InventoryIDCode)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	stmt := `
	select
	serial
	,result
	,note
	from inventory_serial
	where inventory_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, iI)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err2)
		}
	}()

	var iss []InventorySerial

	for rows.Next() {
		var (
			is InventorySerial
		)

		err = rows.Scan(
			&is.Serial.NanoID,
			&is.Result,
			&is.Note,
		)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}

		s, err := db.Serial(is.Serial.NanoID)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}
		is.Serial = *s
		iss = append(iss, is)
	}

	err = rows.Err()
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	return iss, nil
}

func (db *Data) Inventory(id string) (*Inventory, error) {
	iI, err := id64(id, InventoryIDCode)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	stmt := `
	select
	$1||id
	,created_at
	,expected_at
	,started_at
	,ended_at
	,balanced
	,$2||warehouse_id
	,version
	,note
	,$3||account_id
	from inventory
	where id = $4
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	i := new(Inventory)

	err = db.DB.QueryRowContext(ctx, stmt, InventoryIDCode, WarehouseIDCode, AccountIDCode, iI).Scan(
		&i.ID,
		&i.CreatedAt,
		&i.ExpectedAt,
		&i.StartedAt,
		&i.EndedAt,
		&i.Balanced,
		&i.Account.Warehouse.ID,
		&i.Version,
		&i.Note,
		&i.Account.ID,
	)
	if err != nil {
		db.logger.Error(err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w; id: %v", ErrNoInventories, id)
		}
		return nil, err
	}

	a, err := db.Account(i.Account.ID)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}
	i.Account = *a

	i.Items, err = db.InventoryItems(id)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	i.InventorySerials, err = db.InventorySerials(id)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	return i, nil
}

// UncheckedInventorySerials returns an inventory's unchecked serials
func (db *Data) UncheckedInventorySerials(inventoryID string) ([]InventorySerial, error) {
	iss, err := db.InventorySerials(inventoryID)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	var result []InventorySerial

	for _, is := range iss {
		if is.Result == InventoryUnchecked {
			result = append(result, is)
		}
	}

	return result, nil
}

func (db *Data) UncheckedBinByInventory(inventoryID string) (*Bin, error) {
	iI, err := id64(inventoryID, InventoryIDCode)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	stmt := `
	select
	distinct $1||serial.bin_id
	from inventory_serial as isr
	join serial on serial.nanoid = isr.serial
	where isr.result = $2
	and inventory_id = $3
	limit 1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b := new(Bin)

	err = db.DB.QueryRowContext(ctx, stmt, BinIDCode, InventoryUnchecked, iI).Scan(
		&b.ID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			db.logger.Error(fmt.Sprintf("no more unprocessed bin for inventory %v", inventoryID))
			return nil, nil
		}
		db.logger.Error(err.Error())
		return nil, err
	}

	b, err = db.Bin(b.ID)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	return b, nil
}

func (db *Data) UncheckedInventorySerialsOf1RandomBin(inventoryID string) ([]InventorySerial, error) {
	iss, err := db.UncheckedInventorySerials(inventoryID)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	b, err := db.UncheckedBinByInventory(inventoryID)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}
	if b == nil {
		return nil, nil
	}

	var result []InventorySerial

	for _, is := range iss {
		if is.Serial.Bin.ID == b.ID {
			result = append(result, is)
		}
	}

	return result, nil
}

func updateInventoryStartedAt(tx *sql.Tx, inventoryID64 int64) error {
	stmt := `
	update inventory set started_at = now()
	where id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, inventoryID64)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) UpdateInventoryStartedAt(inventoryID string) error {
	iI, err := id64(inventoryID, InventoryIDCode)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}
	defer tx.Rollback()

	err = updateInventoryStartedAt(tx, iI)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	return nil
}

func updateAfterInventoryBinProcessing(tx *sql.Tx, binProcessingResult *Inventory) error {
	iI, err := id64(binProcessingResult.ID, InventoryIDCode)
	if err != nil {
		return err
	}

	stmt := `
	update inventory_serial set
	result = $1
	,note = $2
	where serial = $3
	and inventory_id = $4
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, is := range binProcessingResult.InventorySerials {
		_, err = tx.ExecContext(ctx, stmt, is.Result, is.Note, is.Serial.NanoID, iI)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Data) UpdateAfterInventoryBinProcessing(binProcessingResult *Inventory) error {
	tx, err := db.DB.Begin()
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}
	defer tx.Rollback()

	err = updateAfterInventoryBinProcessing(tx, binProcessingResult)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	return nil
}

func updateInventoryEndedAt(tx *sql.Tx, inventoryID64 int64) error {
	stmt := `
	update inventory set ended_at = now()
	where id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, inventoryID64)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) UpdateInventoryEndedAt(inventoryID string) error {
	iI, err := id64(inventoryID, InventoryIDCode)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}
	defer tx.Rollback()

	err = updateInventoryEndedAt(tx, iI)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	return nil
}

func (db *Data) Inventories(warehouseID string) ([]Inventory, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	stmt := `
	select
	$1||id
	from inventory
	where warehouse_id = $2
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, InventoryIDCode, wI)
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	var is []Inventory

	for rows.Next() {
		i := new(Inventory)

		err = rows.Scan(
			&i.ID,
		)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}

		i, err := db.Inventory(i.ID)
		if err != nil {
			db.logger.Error(err.Error())
			return nil, err
		}

		is = append(is, *i)
	}

	err = rows.Err()
	if err != nil {
		db.logger.Error(err.Error())
		return nil, err
	}

	return is, nil
}

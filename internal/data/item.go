package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/tanNguyen2220022/wh/internal/util"
)

type Item struct {
	GTIN           string  `json:"gtin,omitempty,omitzero"`
	Characteristic string  `json:"characteristic,omitempty,omitzero"`
	Volume         float64 `json:"volume,omitempty,omitzero"`
	Weight         int64   `json:"weight,omitempty,omitzero"`
	Brand          string  `json:"brand,omitempty,omitzero"`
	Material       string  `json:"material,omitempty,omitzero"`
	Color          string  `json:"color,omitempty,omitzero"`
	Size           string  `json:"size,omitempty,omitzero"`
	Price          float32 `json:"price,omitempty,omitzero"`
	Currency       string  `json:"currency,omitempty,omitzero"`
	Type           string  `json:"type,omitempty,omitzero"`
	ShelfLife      int64   `json:"shelfLife,omitempty,omitzero"`
	Img            []byte  `json:"img,omitempty,omitzero"`
	ImgFSPath      string  `json:"imgFSPath,omitempty,omitzero"`
	Name           string  `json:"name,omitempty,omitzero"`
	Stock          int64   `json:"stock,omitempty,omitzero"`
	Version        int     `json:"version,omitempty,omitzero"`
	Supplier       `json:"supplier,omitempty,omitzero"`
}

type ItemQuantity struct {
	Quantity            int64                     `json:"quantity,omitempty,omitzero"`
	ActualQuantity      int64                     `json:"actualQuantity,omitempty,omitzero"`
	MaxReceiveQuantity  int64                     `json:"maxReceiveQuantity,omitempty,omitzero"`
	MaxResupplyQuantity int64                     `json:"maxResupplyQuantity,omitempty,omitzero"`
	Note                string                    `json:"note,omitempty,omitzero"`
	PutawayNote         string                    `json:"putawayNote,omitempty,omitzero"`
	PackNote            string                    `json:"packNote,omitempty,omitzero"`
	PickNote            string                    `json:"pickNote,omitempty,omitzero"`
	Serials             []Serial                  `json:"serials,omitempty,omitzero"`
	Putaway             map[string][]ItemQuantity `json:"putaway,omitempty,omitzero"`
	Item                `json:"item,omitempty,omitzero"`
	Receive             `json:"receive,omitempty,omitzero"`
	Export              `json:"export,omitempty,omitzero"`
}

var (
	ErrNoItems = errors.New("data: no items found")
)

const (
	Shirt      = "Áo sơmi"
	TShirt     = "Áo thun"
	JeanShirt  = "Áo jean"
	WoolShirt  = "Áo len"
	Jacket     = "Áo khoác"
	Jeans      = "Quần jean"
	Trousers   = "Quần tây"
	Sweatpants = "Quần thun"
	Skirt      = "Váy"
	Dress      = "Đầm"
	Onesie     = "Onesie"
)

var (
	selectItemsStmt = fmt.Sprintf(`
	select
	item.gtin,
	brand,
	characteristic,
	color,
	currency,
	img_fspath,
	material,
	price,
	shelf_life,
	size,
	type,
	volume,
	weight,
	type||', '||brand||', màu '||color||', cỡ '||size||', '||characteristic,
	'%v'||supplier_item.supplier_id
	from
	item
	join supplier_item on supplier_item.gtin = item.gtin`, SupplierIDCode)
)

func (db *Data) Item(gtin string) (*Item, error) {
	stmt := selectItemsStmt + "\nwhere supplier_item.gtin=$1"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		i Item
	)
	err := db.DB.QueryRowContext(ctx, stmt, gtin).Scan(
		&i.GTIN,
		&i.Brand,
		&i.Characteristic,
		&i.Color,
		&i.Currency,
		&i.ImgFSPath,
		&i.Material,
		&i.Price,
		&i.ShelfLife,
		&i.Size,
		&i.Type,
		&i.Volume,
		&i.Weight,
		&i.Name,
		&i.Supplier.ID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %v", ErrNoItems, gtin)
	} else if err != nil {
		return nil, err
	}

	return &i, nil
}

func (db *Data) Items(gtins ...string) ([]Item, error) {
	stmt := selectItemsStmt
	if len(gtins) != 0 {
		stmt += "\nwhere gtin in " + Range(int64(len(gtins)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, util.AnySlice(gtins...)...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err2)
		}
	}()

	var is []Item
	for rows.Next() {
		var i Item

		err = rows.Scan(
			&i.GTIN,
			&i.Brand,
			&i.Characteristic,
			&i.Color,
			&i.Currency,
			&i.ImgFSPath,
			&i.Material,
			&i.Price,
			&i.ShelfLife,
			&i.Size,
			&i.Type,
			&i.Volume,
			&i.Weight,
			&i.Name,
			&i.Supplier.ID,
		)
		if err != nil {
			return nil, err
		}

		is = append(is, i)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return is, nil
}

// Stock returns the currently stored and exported quantity of a gtin
func (db *Data) Stock(gtin string, warehouseID string) (int64, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return 0, err
	}

	stmt := `select
	count(*)
	from
	serial
	join purchase on purchase.id = serial.purchase_id
	where gtin=$1
	and receive_tote != 0
	and export_id is null
	and purchase.warehouse_id = $2
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var quantity int64

	err = db.DB.QueryRowContext(ctx, stmt, gtin, wI).Scan(&quantity)
	if err != nil {
		return 0, err
	}

	return quantity, nil
}

func (db *Data) GTINsBySupplier(supplierID string) ([]string, error) {
	i, err := id64(supplierID, SupplierIDCode)
	if err != nil {
		return nil, err
	}

	stmt := `select
	gtin
	from supplier_item
	where supplier_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	var gtins []string
	for rows.Next() {
		var gtin string

		err = rows.Scan(&gtin)
		if err != nil {
			return nil, err
		}

		gtins = append(gtins, gtin)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return gtins, nil
}

func (db *Data) IsGTINBySupplier(gtin, supplierID string) (bool, error) {
	gtins, err := db.GTINsBySupplier(supplierID)
	if err != nil {
		return false, err
	}

	return slices.Contains(gtins, gtin), nil
}

func (db *Data) Volume(is []ItemQuantity) (float64, error) {
	var volume float64
	for _, i := range is {
		it, err := db.Item(i.Item.GTIN)
		if err != nil {
			return 0, err
		}

		volume += it.Volume * float64(i.Quantity)
	}

	return volume, nil
}

func (db *Data) Name(gtin string) (string, error) {
	stmt := `
	select
	type||brand||color||size||characteristic
	from
	item
	where
	gtin = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var name string

	err := db.DB.QueryRowContext(ctx, stmt, gtin).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (db *Data) ItemsBySupplier(supplierID string) ([]Item, error) {
	is, err := db.Items()
	if err != nil {
		return nil, err
	}

	var s []Item
	for _, it := range is {
		if it.Supplier.ID == supplierID {
			s = append(s, it)
		}
	}

	return s, nil
}

func (db *Data) CurrentItemQuantitiesInBinsByWarehouse(warehouseID string) (iqs []ItemQuantity, err error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := `
	select
	serial.gtin
	,type||', '||brand||', màu '||color||', cỡ '||size||', '||characteristic
	,item.img_fspath
	,count(*) as quantity
	from serial
	join purchase on purchase.id = serial.purchase_id
	join item on item.gtin = serial.gtin
	where bin_id is not null and (pick_tote is null and export_id is null)
	and purchase.warehouse_id = $1
	group by item.type, item.brand, item.color, item.size, item.characteristic, item.img_fspath, serial.gtin
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
		var iq ItemQuantity

		err = rows.Scan(
			&iq.Item.GTIN,
			&iq.Item.Name,
			&iq.Item.ImgFSPath,
			&iq.Quantity,
		)
		if err != nil {
			return nil, err
		}

		iqs = append(iqs, iq)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return iqs, nil
}

func (db *Data) StocksByWarehouse(warehouseID string) ([]ItemQuantity, error) {
	currents, err := db.CurrentItemQuantitiesInBinsByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	resupplies, err := db.ResupplyItemsQuantityByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	if len(resupplies) == 0 {
		return currents, nil
	}

	var is []ItemQuantity

	for _, c := range currents {
		for _, r := range resupplies {
			if c.Item.GTIN == r.Item.GTIN {
				c.Quantity -= r.Quantity
				break
			}
		}
		if c.Quantity > 0 {
			is = append(is, c)
		}
	}

	return is, nil
}

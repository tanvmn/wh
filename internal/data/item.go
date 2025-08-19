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
	// Item Item    `json:"item,omitempty,omitzero"`
	Item     `json:"item,omitempty,omitzero"`
	Receive  `json:"receive,omitempty,omitzero"`
	Export   `json:"export,omitempty,omitzero"`
	Quantity int64 `json:"quantity,omitempty,omitzero"`
}

type Serial struct {
	NanoID      string   `json:"nanoID,omitempty,omitzero"`
	ReceiveTote Tote     `json:"receiveTote,omitempty,omitzero"`
	PickTote    Tote     `json:"pickTote,omitempty,omitzero"`
	Bin         Bin      `json:"bin,omitempty,omitzero"`
	Receive     Receive  `json:"receive,omitempty,omitzero"`
	Purchase    Purchase `json:"purchase,omitempty,omitzero"`
	GTIN        string   `json:"gtin,omitempty,omitzero"`
	Export      Export   `json:"export,omitempty,omitzero"`
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

	for i := range is {
		is[i].Stock, err = db.Stock(is[i].GTIN)
		if err != nil && !errors.Is(err, ErrNoItems) {
			return nil, err
		}
	}

	return is, nil
}

// Stock returns the currently stored and exported quantity of a gtin
func (db *Data) Stock(gtin string) (int64, error) {
	stmt := `select
	count(gtin)
	from
	serial
	where
	gtin=$1
	and pick_tote=0
	and export_id=0`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var quantity int64

	err := db.DB.QueryRowContext(ctx, stmt, gtin).Scan(&quantity)
	if err != nil {
		return 0, err
	}

	return quantity, nil
}

// Serials returns the serials of a gtin
func (db *Data) Serials(gtin string) ([]Serial, error) {
	if len(gtin) < 5 {
		return nil, fmt.Errorf("%w: %v", ErrNoItems, gtin)
	}

	stmt := fmt.Sprintf(
		`select
	'%v'||nanoid
	,receive_tote
	,pick_tote
	,bin_id
	,'%v'||bin.warehouse_id
	,bin.shelf
	,bin.row
	,bin.col
	,warehouse.name
	,'%v'||serial.receive_id
	,receive.actual_at
	,'%v'||serial.purchase_id
	,gtin
	,'%v'||export_id
	from
	serial
	join bin on serial.bin_id = bin.id
	join warehouse on bin.warehouse_id = warehouse.id
	join receive on serial.receive_id = receive.id
	where
	gtin = $1`,
		SerialIDCode,
		WarehouseIDCode,
		ReceiveIDCode,
		PurchaseIDCode,
		ExportIDCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, gtin)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err2)
		}
	}()

	ss := []Serial{}
	for rows.Next() {
		var s Serial

		err = rows.Scan(
			&s.NanoID,
			&s.ReceiveTote.ID,
			&s.PickTote.ID,
			&s.Bin.ID,
			&s.Bin.Warehouse.ID,
			&s.Bin.Shelf,
			&s.Bin.Row,
			&s.Bin.Col,
			&s.Bin.Warehouse.Name,
			&s.Receive.ID,
			&s.Receive.ActualAt,
			&s.Purchase.ID,
			&s.GTIN,
			&s.Export.ID,
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

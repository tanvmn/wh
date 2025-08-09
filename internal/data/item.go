package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tanNguyen2220022/wh/internal/util"
)

type Item struct {
	GTIN           string  `json:"gtin,omitempty,omitzero"`
	Characteristic string  `json:"characteristic,omitempty,omitzero"`
	Volume         float32 `json:"volume,omitempty,omitzero"`
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
	Onsie      = "Onsie"
)

const (
	selectItemsStmt = `select
	brand,
	characteristic,
	color,
	currency,
	gtin,
	img_fspath,
	material,
	price,
	shelf_life,
	size,
	type,
	volume,
	weight
	from
	item`
)

func (d *Data) Item(gtin string) (*Item, error) {
	if gtin == "" {
		return nil, ErrNoItems
	}

	stmt := selectItemsStmt + "\nwhere gtin=$1"

	var (
		i Item
	)
	err := d.DB.QueryRow(stmt, gtin).Scan(
		&i.Brand,
		&i.Characteristic,
		&i.Color,
		&i.Currency,
		&i.GTIN,
		&i.ImgFSPath,
		&i.Material,
		&i.Price,
		&i.ShelfLife,
		&i.Size,
		&i.Type,
		&i.Volume,
		&i.Weight,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoItems
	} else if err != nil {
		return nil, err
	}

	return &i, nil
}

func (d *Data) Items(gtins ...string) ([]Item, error) {
	stmt := selectItemsStmt
	if len(gtins) != 0 {
		stmt += "\nwhere gtin in " + Range(int64(len(gtins)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	rows, err := d.DB.QueryContext(ctx, stmt, util.AnySlice(gtins...)...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			d.logger.Error(err.Error())
		}
	}()

	var is []Item
	for rows.Next() {
		var i Item

		err = rows.Scan(
			&i.Brand,
			&i.Characteristic,
			&i.Color,
			&i.Currency,
			&i.GTIN,
			&i.ImgFSPath,
			&i.Material,
			&i.Price,
			&i.ShelfLife,
			&i.Size,
			&i.Type,
			&i.Volume,
			&i.Weight,
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
		is[i].Stock, err = d.Stock(is[i].GTIN)
		if err != nil && !errors.Is(err, ErrNoItems) {
			return nil, err
		}
	}

	return is, nil
}

// Stock returns the currently stored and exported quantity of a gtin
func (d *Data) Stock(gtin string) (int64, error) {
	stmt := `select
	count(gtin)
	from
	seri
	where
	gtin=$1
	and pick_tote=0
	and export_id=0`

	var quantity int64
	err := d.DB.QueryRow(stmt, gtin).Scan(&quantity)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoItems
	} else if err != nil {
		return 0, err
	}

	return quantity, nil
}

// Serials returns the serials of a gtin
func (d *Data) Serials(gtin string) ([]Serial, error) {
	stmt := `select
	'SER-'||nanoid
	,receive_tote
	,pick_tote
	,bin_id
	,'WAR'||bin.warehouse_id
	,bin.shelf
	,bin.row
	,bin.col
	,warehouse.name
	,'REC'||serial.receive_id
	,receive.actual_dtime
	,'PUR'||serial.purchase_id
	,gtin
	,'EXP'||export_id
	from
	serial
	join bin on serial.bin_id = bin.id
	join warehouse on bin.warehouse_id = warehouse.id
	join receive on serial.receive_id = receive.id
	where
	gtin = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	rows, err := d.DB.QueryContext(ctx, stmt, gtin)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			d.logger.Error(err.Error())
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

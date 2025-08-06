package data

import (
	"database/sql"
	"errors"
)

const (
	Shirt      = "Áo sơmi"
	TShirt     = "Áo thun"
	JeanShirt  = "Áo jean"
	WoolShirt  = "Áo len"
	Jeans      = "Quần jean"
	Trousers   = "Quần tây"
	Sweatpants = "Quần thun"
	Skirt      = "Váy"
	Dress      = "Đầm"
	Onsie      = "Onsie"
)

type Item struct {
	Brand          string  `json:"brand,omitempty,omitzero"`
	Characteristic string  `json:"characteristic,omitempty,omitzero"`
	Color          string  `json:"color,omitempty,omitzero"`
	Currency       string  `json:"currency,omitempty,omitzero"`
	GTIN           string  `json:"gtin,omitempty,omitzero"`
	Img            []byte  `json:"img,omitempty,omitzero"`
	ImgPath        string  `json:"imgPath,omitempty,omitzero"`
	Material       string  `json:"material,omitempty,omitzero"`
	Name           string  `json:"name,omitempty,omitzero"`
	Price          float32 `json:"price,omitempty,omitzero"`
	ShelfLife      int64   `json:"shelfLife,omitempty,omitzero"`
	Size           string  `json:"size,omitempty,omitzero"`
	Type           string  `json:"type,omitempty,omitzero"`
	Volume         float32 `json:"volume,omitempty,omitzero"`
	Weight         int64   `json:"weight,omitempty,omitzero"`
}

var (
	ErrNoItems = errors.New("data: no items found")
)

func (d *Data) Item(gtin string) (*Item, error) {
	if gtin == "" {
		return nil, ErrNoItems
	}

	stmt := `select
	brand,
	characteristic,
	color,
	currency,
	gtin,
	material,
	price,
	shelf_life,
	size,
	type,
	volume,
	weight
	from item
	where gtin=$1`

	var (
		i Item
	)
	err := d.DB.QueryRow(stmt, gtin).Scan(
		&i.Brand,
		&i.Characteristic,
		&i.Color,
		&i.Currency,
		&i.GTIN,
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

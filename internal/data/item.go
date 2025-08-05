package data

import (
	"database/sql"
	"errors"
)

type Type struct {
	ID   int64  `json:"id,omitempty,omitzero"`
	Name string `json:"name,omitempty,omitzero"`
}

type Item struct {
	GTIN           string  `json:"gtin,omitempty,omitzero"`
	Characteristic string  `json:"characteristic,omitempty,omitzero"`
	Volume         float32 `json:"volume,omitempty,omitzero"`
	Weight         float32 `json:"weight,omitempty,omitzero"`
	Brand          string  `json:"brand,omitempty,omitzero"`
	Material       string  `json:"material,omitempty,omitzero"`
	Color          string  `json:"color,omitempty,omitzero"`
	Size           string  `json:"size,omitempty,omitzero"`
	Price          float32 `json:"price,omitempty,omitzero"`
	Name           string  `json:"name,omitempty,omitzero"`
	Type           `json:"type,omitempty,omitzero"`
}

var (
	ErrNoItems = errors.New("data: no items found")
)

func (d *Data) Item(gtin string) (*Item, error) {
	if gtin == "" {
		return nil, ErrNoItems
	}

	stmt := `select
	gtin,
	characteristic,
	volume,
	weight,
	brand,
	material,
	color,
	size,
	price,
	type_id,
	type.name
	from item
	join type on type_id=type.id
	where gtin=$1`

	var (
		i Item
	)
	err := d.DB.QueryRow(stmt, gtin).Scan(
		&i.GTIN,
		&i.Characteristic,
		&i.Volume,
		&i.Weight,
		&i.Brand,
		&i.Material,
		&i.Color,
		&i.Size,
		&i.Price,
		&i.Type.ID,
		&i.Type.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoItems
	} else if err != nil {
		return nil, err
	}

	return &i, nil
}

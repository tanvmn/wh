package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Warehouse struct {
	ID      string `json:"id,omitempty,omitzero"`
	Name    string `json:"name,omitempty,omitzero"`
	Address string `json:"address,omitempty,omitzero"`
	Phone   string `json:"phone,omitempty,omitzero"`
	Email   string `json:"email,omitempty,omitzero"`
	Version int    `json:"version,omitempty,omitzero"`
}

type Bin struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
	Capacity  float64   `json:"capacity,omitempty,omitzero"`
	Shelf     int64     `json:"shelf,omitempty,omitzero"`
	Row       int64     `json:"row,omitempty,omitzero"`
	Col       int64     `json:"col,omitempty,omitzero"`
	Version   int       `json:"version,omitempty,omitzero"`
}

type Tote struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
	Capacity  float64   `json:"capacity,omitempty,omitzero"`
	Version   int       `json:"version,omitempty,omitzero"`
}

type Store struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Name      string    `json:"name,omitempty,omitzero"`
	Address   string    `json:"address,omitempty,omitzero"`
	Phone     string    `json:"phone,omitempty,omitzero"`
	Email     string    `json:"email,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
	Version   int       `json:"version,omitempty,omitzero"`
}

var (
	ErrNoWarehouses = errors.New("data: no warehouses found")
	ErrNoStores     = errors.New("data: no stores found")
)

func (db *Data) Warehouse(id string) (*Warehouse, error) {
	i, err := id64(id, WarehouseIDCode)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrNoWarehouses, id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(
		`select
	'%v'||id
	,name
	,address
	,phone
	,email
	from warehouse
	where
	id=$1`, WarehouseIDCode)
	var wh Warehouse

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
		&wh.ID,
		&wh.Name,
		&wh.Address,
		&wh.Phone,
		&wh.Email,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w, %v", ErrNoWarehouses, id)
	} else if err != nil {
		return nil, err
	}

	return &wh, nil
}

func (db *Data) Capacity(warehouseID string) (float64, error) {
	i, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return 0, err
	}

	stmt := `
	select
	sum(capacity)
	from bin
	where warehouse_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var capacity sql.NullFloat64

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(&capacity)
	if err != nil {
		return 0, err
	}

	return capacity.Float64, nil
}

func (db *Data) UsedVolume(warehouseID string) (float64, error) {
	i, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return 0, err
	}

	stmt := `
	select
        sum(volume)
        from purchase_item as pi
        join purchase on purchase.id = pi.purchase_id
        join item on item.gtin = pi.gtin
        where not purchase.status in ($1)
	and warehouse_id = $2;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var usedV sql.NullFloat64

	err = db.DB.QueryRowContext(ctx, stmt, Declined, i).Scan(&usedV)
	if err != nil {
		return 0, err
	}

	return usedV.Float64, nil
}

func (db *Data) Warehouses() ([]Warehouse, error) {
	stmt := fmt.Sprintf(`
	select
	'%v'||id
	,name
	,address
	,phone
	,version
	,email
	from warehouse
	`, WarehouseIDCode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var ws []Warehouse

	rows, err := db.DB.QueryContext(ctx, stmt)
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
		var w Warehouse

		err = rows.Scan(
			&w.ID,
			&w.Name,
			&w.Address,
			&w.Phone,
			&w.Version,
			&w.Email,
		)
		if err != nil {
			return nil, err
		}

		ws = append(ws, w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ws, nil
}

func (db *Data) UnusedTotes(warehouseID string) ([]Tote, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	select
	'%v'||tote.id
	,capacity
	,warehouse_id
	,tote.version
	from tote
	left join serial on (serial.receive_tote = tote.id and serial.bin_id is null) or serial.pick_tote = tote.id
	left join export on export.id = serial.export_id and export.packed_at = '1000-01-01'
	where tote.warehouse_id = $1
	and serial.nanoid is null
	;`,
		ToteIDCode,
	)

	rows, err := db.DB.QueryContext(ctx, stmt, wI)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := rows.Close(); err2 != nil {
			panic(err)
		}
	}()

	var ts []Tote

	for rows.Next() {
		var t Tote

		err = rows.Scan(
			&t.ID,
			&t.Capacity,
			&t.Warehouse.ID,
			&t.Version,
		)
		if err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ts, nil
}

func (db *Data) CurrentBinsCapaciy(warehouseID string) ([]Bin, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select
	'%v'||bin.id
	,bin.capacity - case
	when sum(item.volume) is null then 0
	else sum(item.volume)
	end
	from bin
	left join serial on serial.bin_id = bin.id and serial.pick_tote is null
	left join item on item.gtin = serial.gtin
	where bin.warehouse_id = $1
	group by bin.id
	order by bin.id
	;`,
		BinIDCode,
	)

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

	var (
		bs  []Bin
		cap sql.NullFloat64
	)
	for rows.Next() {
		var b Bin
		err = rows.Scan(
			&b.ID,
			&cap,
		)
		if err != nil {
			return nil, err
		}

		if cap.Valid {
			b.Capacity = cap.Float64
		}

		bs = append(bs, b)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (db *Data) Bin(id string) (*Bin, error) {
	i, err := id64(id, BinIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select
	'%v'||id
	,'%v'||warehouse_id
	,shelf
	,row
	,col
	,version
	,capacity
	from bin
	where id = $1
	;`,
		BinIDCode,
		WarehouseIDCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var b Bin
	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
		&b.ID,
		&b.Warehouse.ID,
		&b.Shelf,
		&b.Row,
		&b.Col,
		&b.Version,
		&b.Capacity,
	)
	if err != nil {
		return nil, err
	}

	w, err := db.Warehouse(b.Warehouse.ID)
	if err != nil {
		return nil, err
	}
	b.Warehouse = *w

	return &b, nil
}

func (db *Data) Store(id string) (*Store, error) {
	i, err := id64(id, StoreIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select
	'%v'||id
	,name
	,address
	,email
	,phone
	,version
	,'%v'||warehouse_id
	from store
	where id = $1
	;`,
		StoreIDCode,
		WarehouseIDCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var s Store

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
		&s.ID,
		&s.Name,
		&s.Address,
		&s.Email,
		&s.Phone,
		&s.Version,
		&s.Warehouse.ID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoStores
		}
		return nil, err
	}

	w, err := db.Warehouse(s.Warehouse.ID)
	if err != nil {
		return nil, err
	}
	s.Warehouse = *w

	return &s, nil
}

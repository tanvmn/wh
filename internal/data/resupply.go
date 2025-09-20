package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Resupply struct {
	ID         string         `json:"id,omitempty,omitzero"`
	Status     string         `json:"status,omitempty,omitzero"`
	ExpectedAt string         `json:"expectedAt,omitempty,omitzero"`
	CreatedAt  string         `json:"createdAt,omitempty,omitzero"`
	Note       string         `json:"note,omitempty,omitzero"`
	Version    int            `json:"version,omitempty,omitzero"`
	Items      []ItemQuantity `json:"items,omitempty,omitzero"`
	Account    `json:"account,omitempty,omitzero"`
}

var (
	ErrNoResupplies = errors.New("no resupplies found")
)

func (db *Data) ResupplyItemsQuantityByWarehouse(warehouseID string) ([]ItemQuantity, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select
	ri.gtin
	,type||', '||brand||', màu '||color||', cỡ '||size||', '||characteristic
	,item.img_fspath
	,sum(ri.quantity)
	from resupply_item as ri
	join item on item.gtin = ri.gtin
	join resupply on resupply.id = ri.resupply_id
	join store on store.id = resupply.store_id
	where resupply.status != '%v'
	and store.warehouse_id = $1
	group by item.type, item.brand, item.color, item.size, item.characteristic, item.img_fspath, ri.gtin
	;`,
		Ended,
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

	var iqs []ItemQuantity

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

func addResupply(tx *sql.Tx, r *Resupply) (id string, err error) {
	aI, err := id64(r.Account.ID, AccountIDCode)
	if err != nil {
		return "", err
	}
	sI, err := id64(r.Store.ID, StoreIDCode)
	if err != nil {
		return "", err
	}

	stmt := fmt.Sprintf(`
	insert into resupply (expected_at, account_id, store_id) values
	($1, $2, $3)
	returning '%v'||id
	;`,
		ResupplyIDCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tx.QueryRowContext(ctx, stmt, r.ExpectedAt, aI, sI).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

// Remember r has to already be assigned with the newly created resupply id
func addResupplyItems(tx *sql.Tx, r *Resupply) error {
	rI, err := id64(r.ID, ResupplyIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	insert into resupply_item  (resupply_id, gtin, quantity) values
	($1, $2, $3)
	;`

	for _, iq := range r.Items {
		_, err = tx.ExecContext(ctx, stmt, rI, iq.Item.GTIN, iq.Quantity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Data) AddResupply(r *Resupply) (id string, err error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	r.ID, err = addResupply(tx, r)
	if err != nil {
		return "", err
	}

	err = addResupplyItems(tx, r)
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return r.ID, nil
}

// SetMaxResupplyItemQuantities queries the stocks of the resupply's target warehouse,
// and for each common item that the resupply has with the stocks, sum the 2 quantity
// to set the max quantity that the common item of the resupply could have
func (db *Data) SetMaxResupplyItemQuantities(r *Resupply) error {
	ss, err := db.StocksByWarehouse(r.Account.Store.Warehouse.ID)
	if err != nil {
		return err
	}

	for i := range r.Items {
		for _, s := range ss {
			if r.Items[i].Item.GTIN == s.Item.GTIN {
				r.Items[i].MaxResupplyQuantity = r.Items[i].Quantity + s.Quantity
				break
			}
		}
	}

	return nil
}

func (db *Data) ResupplyItems(resupplyID int64) ([]ItemQuantity, error) {
	stmt := `
	select
	ri.gtin
	,type||', '||brand||', màu '||color||', cỡ '||size||', '||characteristic
	,item.img_fspath
	,ri.quantity
	from resupply_item as ri
	join item on item.gtin = ri.gtin
	where ri.resupply_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, resupplyID)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err)
		}
	}()

	var iqs []ItemQuantity

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

func (db *Data) Resupply(id string) (*Resupply, error) {
	i, err := id64(id, ResupplyIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select
	'%v'||id
	,expected_at
	,to_char(created_at, 'DD-MM-YYYY HH24:MI')
	,status
	,note
	,'%v'||account_id
	,'%v'||store_id
	,version
	from resupply
	where id = $1
	;`,
		ResupplyIDCode,
		AccountIDCode,
		StoreIDCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var r Resupply
	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
		&r.ID,
		&r.ExpectedAt,
		&r.CreatedAt,
		&r.Status,
		&r.Note,
		&r.Account.ID,
		&r.Account.Store.ID,
		&r.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoResupplies
		}
		return nil, err
	}

	r.Items, err = db.ResupplyItems(i)
	if err != nil {
		return nil, err
	}

	a, err := db.Account(r.Account.ID)
	if err != nil {
		return nil, err
	}
	s, err := db.Store(r.Account.Store.ID)
	if err != nil {
		return nil, err
	}
	a.Store = *s
	r.Account = *a

	return &r, nil
}

func setResupply(tx *sql.Tx, r *Resupply) error {
	rI, err := id64(r.ID, ResupplyIDCode)
	if err != nil {
		return err
	}

	stmt := `
	update resupply set expected_at = $1
	, version = version + 1
	where id = $2
	and version = $3
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := tx.ExecContext(ctx, stmt, r.ExpectedAt, rI, r.Version)
	if err != nil {
		return err
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if ra != 1 {
		return ErrSetConflict
	}

	return nil
}

func delResupplyItems(tx *sql.Tx, resupplyID int64) error {
	stmt := `
	delete from resupply_item where resupply_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, resupplyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) SetResupply(r *Resupply) error {
	rI, err := id64(r.ID, ResupplyIDCode)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = delResupplyItems(tx, rI)
	if err != nil {
		return err
	}

	err = setResupply(tx, r)
	if err != nil {
		return err
	}

	err = addResupplyItems(tx, r)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func delResupply(tx *sql.Tx, id int64) error {
	stmt := `
	delete from resupply where id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := tx.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) DelResupply(id string) error {
	i, err := id64(id, ResupplyIDCode)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = delResupplyItems(tx, i)
	if err != nil {
		return err
	}

	err = delResupply(tx, i)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) ResuppliesByWarehouse(warehouseID string) ([]Resupply, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select 
	'%v'||resupply.id
	from resupply
	join store on store.id = resupply.store_id
	where store.warehouse_id = $1
	;`,
		ResupplyIDCode,
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

	var rs []Resupply

	for rows.Next() {
		var r Resupply

		err = rows.Scan(
			&r.ID,
		)
		if err != nil {
			return nil, err
		}

		rs = append(rs, r)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for i := range rs {
		rp, err := db.Resupply(rs[i].ID)
		if err != nil {
			return nil, err
		}
		rs[i] = *rp
	}

	return rs, nil
}

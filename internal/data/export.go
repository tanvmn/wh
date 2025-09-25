package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Export struct {
	ID         string         `json:"id,omitempty,omitzero"`
	CreatedAt  string         `json:"createdAt,omitempty,omitzero"`
	ExpectedAt string         `json:"expectedAt,omitempty,omitzero"`
	PickedAt   string         `json:"pickedAt,omitempty,omitzero"`
	PackedAt   string         `json:"packedAt,omitempty,omitzero"`
	PickedBy   Account        `json:"pickedBy,omitempty,omitzero"`
	PackedBy   Account        `json:"packedBy,omitempty,omitzero"`
	Note       string         `json:"note,omitempty,omitzero"`
	PickNote   string         `json:"pickNote,omitempty,omitzero"`
	PackNote   string         `json:"packNote,omitempty,omitzero"`
	VoucherID  string         `json:"voucherID,omitempty,omitzero"`
	Version    int            `json:"version,omitempty,omitzero"`
	Items      []ItemQuantity `json:"items,omitempty,omitzero"`
	Account    `json:"account,omitempty,omitzero"`
	Transfer   `json:"transfer,omitempty,omitzero"`
	Resupply   `json:"resupply,omitempty,omitzero"`
}

var (
	ErrNoExports = errors.New("no exports found")
)

func delExportItemsByResupply(tx *sql.Tx, resupplyID int64) error {
	stmt := `
	delete from export_item where resupply_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, resupplyID)
	if err != nil {
		return err
	}

	return nil
}

func delExportByResupply(tx *sql.Tx, resupplyID int64) error {
	stmt := `
	delete from export where resupply_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, resupplyID)
	if err != nil {
		return err
	}

	return nil
}

func delExportItems(tx *sql.Tx, exportID int64) error {
	stmt := `
	delete from export_item where export_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, exportID)
	if err != nil {
		return err
	}

	return nil
}

func delExport(tx *sql.Tx, exportID int64) error {
	stmt := `
	delete from export where export_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, exportID)
	if err != nil {
		return err
	}

	return nil
}

func addExport(tx *sql.Tx, r *Resupply) (id int64, err error) {
	aI, err := id64(r.Account.ID, AccountIDCode)
	if err != nil {
		return 0, err
	}
	rI, err := id64(r.ID, ResupplyIDCode)
	if err != nil {
		return 0, err
	}
	voucherID := fmt.Sprintf("%v%v", VoucherIDCode, rI)

	stmt := `
	insert into export (account_id, expected_at, voucher_id, resupply_id) values
	($1, $2, $3, $4)
	returning id
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tx.QueryRowContext(ctx, stmt, aI, r.ExpectedAt, voucherID, rI).Scan(
		&id,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func addExportItems(tx *sql.Tx, resupplyID, exportID int64, is []ItemQuantity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	insert into export_item (export_id, resupply_id, gtin, quantity) values
	($1, $2, $3, $4)
	;`

	for _, i := range is {
		_, err := tx.ExecContext(ctx, stmt, exportID, resupplyID, i.Item.GTIN, i.Quantity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Data) AddExport(resupplyID string) (exportID string, err error) {
	rI, err := id64(resupplyID, ResupplyIDCode)
	if err != nil {
		return "", err
	}

	r, err := db.Resupply(resupplyID)
	if err != nil {
		return "", err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	exportID64, err := addExport(tx, r)
	if err != nil {
		return "", err
	}

	err = addExportItems(tx, rI, exportID64, r.Items)
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v%v", ExportIDCode, exportID64), nil
}

func (db *Data) ExportItems(resupplyID, exportID int64) ([]ItemQuantity, error) {
	stmt := `
	select
	ei.gtin
	,ei.note
	,ei.pick_note
	,ei.pack_note
	,ei.quantity
	from export_item as ei
	where export_id = $1
	and resupply_id = $2
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var is []ItemQuantity

	rows, err := db.DB.QueryContext(ctx, stmt, exportID, resupplyID)
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
			&iq.Note,
			&iq.PickNote,
			&iq.PackNote,
			&iq.Quantity,
		)
		if err != nil {
			return nil, err
		}

		i, err := db.Item(iq.Item.GTIN)
		if err != nil {
			return nil, err
		}
		iq.Item = *i

		is = append(is, iq)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return is, nil
}

func (db *Data) Export(id string) (*Export, error) {
	i, err := id64(id, ExportIDCode)
	if err != nil {
		return nil, err
	}

	stmt := `
	select
	$1||export.id
	,$2||export.account_id
	,$2||export.picked_by
	,$2||export.packed_by
	,to_char(export.created_at, 'DD-MM-YYYY HH24:MI')
	,export.expected_at
	,export.picked_at
	,export.packed_at
	,export.note
	,export.pick_note
	,export.pack_note
	,export.voucher_id
	,$3||export.resupply_id
	,$4||export.transfer_id
	,export.version
	from export
	where export.id = $5
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		pickedBy, packedBy, transferID sql.NullString
		e                              Export
	)

	err = db.DB.QueryRowContext(ctx, stmt,
		ExportIDCode,
		AccountIDCode,
		ResupplyIDCode,
		TransferIDCode,
		i,
	).Scan(
		&e.ID,
		&e.Account.ID,
		&pickedBy,
		&packedBy,
		&e.CreatedAt,
		&e.ExpectedAt,
		&e.PickedAt,
		&e.PackedAt,
		&e.Note,
		&e.PickNote,
		&e.PackNote,
		&e.VoucherID,
		&e.Resupply.ID,
		&transferID,
		&e.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w; id %v", ErrNoExports, id)
		}
		return nil, err
	}

	if pickedBy.Valid {
		a, err := db.Account(pickedBy.String)
		if err != nil {
			return nil, err
		}
		e.PickedBy = *a
	}

	if packedBy.Valid {
		a, err := db.Account(packedBy.String)
		if err != nil {
			return nil, err
		}
		e.PackedBy = *a
	}

	{
		a, err := db.Account(e.Account.ID)
		if err != nil {
			return nil, err
		}
		e.Account = *a
	}

	r, err := db.Resupply(e.Resupply.ID)
	if err != nil {
		return nil, err
	}
	e.Resupply = *r

	// Get the export's items
	rI, err := id64(e.Resupply.ID, ResupplyIDCode)
	if err != nil {
		return nil, err
	}
	e.Items, err = db.ExportItems(rI, i)
	if err != nil {
		return nil, err
	}

	if transferID.Valid {
		e.Transfer.ID = transferID.String
	}

	return &e, nil
}

func (db *Data) ExportsByWarehouse(warehouseID string) ([]Export, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := `
	select
	$1||export.id
	from export
	join resupply on resupply.id = export.resupply_id
	join store on store.id = resupply.store_id
	where store.warehouse_id = $2
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, ExportIDCode, wI)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := rows.Close(); err2 != nil {
			panic(err2)
		}
	}()

	var es []Export

	for rows.Next() {
		e := new(Export)

		err = rows.Scan(
			&e.ID,
		)
		if err != nil {
			return nil, err
		}

		e, err := db.Export(e.ID)
		if err != nil {
			return nil, err
		}

		es = append(es, *e)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return es, nil
}

func (db *Data) CalculatedPicks(exportID string) ([]ItemQuantity, error) {
	e, err := db.Export(exportID)
	if err != nil {
		return nil, err
	}

	bi, err := db.BinAndItemQuantityByWarehouse(e.Account.Store.Warehouse.ID)
	if err != nil {
		return nil, err
	}

	var is []ItemQuantity

	// for each export item:
	//   for each bin that has the stock of the export item:
	//     if export item's quantity <= stock's quantity in bin, then pick all from that bin and move onto the next export item
	//     else pick all from the bin, calculate the remain quantity of that export item and move onto the next bin that has stock of the export item
	for _, ei := range e.Items {
		for i := range bi {
			if ei.Item.GTIN == bi[i].Item.GTIN {
				var iq ItemQuantity
				iq.Item = ei.Item
				iq.PickBin = bi[i].PickBin

				if ei.Quantity <= bi[i].Quantity {
					iq.Quantity = ei.Quantity
					bi[i].Quantity -= ei.Quantity

					is = append(is, iq)
					break
				}
				iq.Quantity = ei.Quantity - bi[i].Quantity
				ei.Quantity -= bi[i].Quantity
				bi[i].Quantity = 0

				is = append(is, iq)
			}
		}
	}

	return is, nil
}

// updateExportAfterPick set an export's pick_by, pick_at based on pick result from client
func updateExportAfterPick(tx *sql.Tx, pickResult *Export) error {
	aI, err := id64(pickResult.PickedBy.ID, AccountIDCode)
	if err != nil {
		return err
	}
	eI, err := id64(pickResult.ID, ExportIDCode)
	if err != nil {
		return err
	}

	stmt := `
	update export
	set picked_by = $1
	,picked_at = now()
	,version = version + 1
	where id = $2
	and version = $3
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := tx.ExecContext(ctx, stmt, aI, eI, pickResult.Version)
	if err != nil {
		return err
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if ra != 1 {
		return fmt.Errorf("%w; update after pick, export %v, verison %v", ErrSetConflict, pickResult.ID, pickResult.Version)
	}

	return nil
}

func updateExportItemAfterPick(tx *sql.Tx, pickResult *Export) error {
	eI, err := id64(pickResult.ID, ExportIDCode)
	if err != nil {
		return err
	}
	rI, err := id64(pickResult.Resupply.ID, ResupplyIDCode)
	if err != nil {
		return err
	}

	stmt := `
	update export_item
	set pick_note = $1
	where export_id = $2
	and resupply_id = $3
	and gtin = $4
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, iq := range pickResult.Items {
		if iq.PickNote != "" {
			_, err = tx.ExecContext(ctx, stmt, iq.PickNote, eI, rI, iq.Item.GTIN)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *Data) PickExport(pickResult *Export) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = updateExportAfterPick(tx, pickResult)
	if err != nil {
		return err
	}

	err = updateExportItemAfterPick(tx, pickResult)
	if err != nil {
		return err
	}

	err = updateSerialAfterPick(tx, pickResult)
	if err != nil {
		return err
	}

	rI, err := id64(pickResult.Resupply.ID, ResupplyIDCode)
	if err != nil {
		return err
	}
	err = updateResupplyAfterPick(tx, rI, pickResult.Resupply.Version)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

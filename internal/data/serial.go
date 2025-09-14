package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Serial struct {
	NanoID      string `json:"nanoID,omitempty,omitzero"`
	GTIN        string `json:"gtin,omitempty,omitzero"`
	ReceiveTote Tote   `json:"receiveTote,omitempty,omitzero"`
	PickTote    Tote   `json:"pickTote,omitempty,omitzero"`
	Bin         `json:"bin,omitempty,omitzero"`
	Purchase    `json:"purchase,omitempty,omitzero"`
	Receive     `json:"receive,omitempty,omitzero"`
	Resupply    `json:"resupply,omitempty,omitzero"`
	Export      `json:"export,omitempty,omitzero"`
}

// Serials returns the serials of a gtin
func (db *Data) SerialsByGTINAndWarehouse(gtin string, warehouseID string) ([]Serial, error) {
	if len(gtin) < 5 {
		return nil, fmt.Errorf("%w: %v", ErrNoItems, gtin)
	}
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(
		`select
		nanoid
		,'%v'||receive_tote
		,'%v'||pick_tote
		,'%v'||bin_id
		,bin.shelf
		,bin.row
		,bin.col
		,'%v'||serial.receive_id
		,to_char(receive.actual_at, 'DD-MM-YYYY HH24:MI')
		,'%v'||serial.purchase_id
		,'%v'||purchase.warehouse_id
		,warehouse.name
		,gtin
		from
		serial
		join receive on serial.receive_id = receive.id
		join purchase on serial.purchase_id = purchase.id
		join warehouse on purchase.warehouse_id = warehouse.id
		left join export on export.id = serial.export_id
		left join bin on serial.bin_id = bin.id
		where gtin = $1
		and (export.picked_at = '1000-01-01 00:00' or export.picked_at is null)
		and warehouse.id = $2
		;`,
		ToteIDCode,
		ToteIDCode,
		BinIDCode,
		ReceiveIDCode,
		PurchaseIDCode,
		WarehouseIDCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, gtin, wI)
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
		var (
			s                        Serial
			binRow, binCol, binShelf sql.NullInt64
			// pickToteID, binID, exportID sql.NullString
			pickToteID, binID sql.NullString
		)

		err = rows.Scan(
			&s.NanoID,
			&s.ReceiveTote.ID,
			&pickToteID,
			&binID,
			&binShelf,
			&binRow,
			&binCol,
			&s.Receive.ID,
			&s.Receive.ActualAt,
			&s.Purchase.ID,
			&s.Purchase.Warehouse.ID,
			&s.Purchase.Warehouse.Name,
			&s.GTIN,
			// &exportID,
		)
		if err != nil {
			return nil, err
		}

		if pickToteID.Valid {
			s.PickTote.ID = pickToteID.String
		}

		if binID.Valid {
			s.Bin.ID = binID.String
		}
		if binShelf.Valid {
			s.Bin.Shelf = binShelf.Int64
		}
		if binRow.Valid {
			s.Bin.Row = binRow.Int64
		}
		if binCol.Valid {
			s.Bin.Col = binCol.Int64
		}

		// if exportID.Valid {
		// 	s.Export.ID = exportID.String
		// }

		ss = append(ss, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ss, nil
}

func addSerial(tx *sql.Tx, s *Serial) error {
	rID, err := id64(s.Receive.ID, ReceiveIDCode)
	if err != nil {
		return err
	}
	pID, err := id64(s.Purchase.ID, PurchaseIDCode)
	if err != nil {
		return err
	}
	tID, err := id64(s.ReceiveTote.ID, ToteIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	insert into serial (nanoid, gtin, receive_tote, receive_id, purchase_id) values
	($1,$2,$3,$4,$5)
	;`

	// println("rID", rID)
	// println("pID", pID)
	// println("tID", tID)
	// println(s.GTIN)
	_, err = tx.ExecContext(ctx, stmt, s.NanoID, s.GTIN, tID, rID, pID)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) AddSerial(s *Serial) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = addSerial(tx, s)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) AddSerials(ss []Serial) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, s := range ss {
		err = addSerial(tx, &s)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) SerialsByReceive(rc *Receive) error {
	rI, err := id64(rc.ID, ReceiveIDCode)
	if err != nil {
		return err
	}
	wI, err := id64(rc.Purchase.Warehouse.ID, WarehouseIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	select
	nanoid
	,serial.gtin
	,'%v'||serial.bin_id
	,shelf
	,row
	,col
	,'%v'||receive_tote
	,'%v'||pick_tote
	,'%v'||serial.purchase_id
	,'%v'||serial.receive_id
	,'%v'||serial.export_id
	from serial
	join receive on receive.id = serial.receive_id
	join purchase on purchase.id = receive.purchase_id
	left join bin on bin.id = serial.bin_id
	where serial.receive_id = $1
	and purchase.warehouse_id = $2
	;`,
		BinIDCode,
		ToteIDCode,
		ToteIDCode,
		PurchaseIDCode,
		ReceiveIDCode,
		ExportIDCode,
	)

	rows, err := db.DB.QueryContext(ctx, stmt, rI, wI)
	if err != nil {
		return err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err2)
		}
	}()

	var (
		ss                        []Serial
		binID, pickTote, exportID sql.NullString
		shelf, row, col           sql.NullInt64
	)

	for rows.Next() {
		var s Serial
		err = rows.Scan(
			&s.NanoID,
			&s.GTIN,
			&binID,
			&shelf,
			&row,
			&col,
			&s.ReceiveTote.ID,
			&pickTote,
			&s.Purchase.ID,
			&s.Receive.ID,
			&exportID,
		)
		if err != nil {
			return err
		}

		if shelf.Valid {
			s.Bin.Shelf = shelf.Int64
		}
		if row.Valid {
			s.Bin.Row = row.Int64
		}
		if col.Valid {
			s.Bin.Col = col.Int64
		}

		if binID.Valid {
			s.Bin.ID = binID.String
		}

		if pickTote.Valid {
			s.PickTote.ID = pickTote.String
		}

		if exportID.Valid {
			s.Export.ID = exportID.String
		}

		ss = append(ss, s)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	for i := range rc.Items {
		for _, s := range ss {
			if rc.Items[i].GTIN == s.GTIN {
				rc.Items[i].Serials = append(rc.Items[i].Serials, s)
			}
		}
	}

	return nil
}

func delUnputawaySerials(tx *sql.Tx, rc *Receive) error {
	stmt := `
	delete from serial where nanoid = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, iq := range rc.Items {
		for _, s := range iq.Serials {
			if len(s.Bin.ID) == 0 && len(iq.PutawayNote) != 0 {
				_, err := tx.ExecContext(ctx, stmt, s.NanoID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (db *Data) DelUnputawaySerials(rc *Receive) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = delUnputawaySerials(tx, rc)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// ExpectedPutawaySerialsByPutawayReceive adds the expected serials to putaway by receiveID
func (db *Data) AddDifferenceSerialsByGTINOfPutawayReceive(rc *Receive) error {
	for i := range rc.Items {
		ss, err := db.DifferenceSerialsByGTINOfPutawayReceive(rc.Purchase.Warehouse.ID, rc.ID, rc.Items[i].GTIN)
		if err != nil {
			return err
		}

		rc.Items[i].Serials = append(rc.Items[i].Serials, ss...)
	}

	return nil
}

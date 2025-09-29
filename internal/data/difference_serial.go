package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// AddDiffrenceSerials takes a *Receive that WAS USED TO CATCH JSON from the putaway result
// and add the difference serials if exist
func (db *Data) AddPutawayDifferenceSerials(putawayResult *Receive) error {
	for _, iq := range putawayResult.Items {
		for _, s := range iq.Serials {
			if len(s.Bin.ID) == 0 && len(iq.PutawayNote) != 0 {
				stmt := `
				insert into difference_serial (activity_id, nanoid, receive_tote, receive_id, purchase_id, gtin) values
				($1
				,$2
				,substring($3 from 5 for 1)::bigint
				,substring($4 from 5 for 1)::bigint
				,substring($5 from 5 for 1)::bigint
				,$6
				)
				;`

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				_, err := db.DB.ExecContext(ctx, stmt,
					PutawayIDCode+putawayResult.ID[4:],
					s.NanoID,
					s.ReceiveTote.ID,
					s.Receive.ID,
					s.Purchase.ID,
					s.GTIN,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// DifferenceSerialsByWarehouse returns all difference serials of warehouseID
func (db *Data) DifferenceSerialsByWarehouse(warehouseID string) ([]Serial, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select
	nanoid
	,'%v'||receive_tote
	,'%v'||pick_tote
	,'%v'||bin_id
	,'%v'||receive_id
	,'%v'||purchase_id
	,gtin
	,'%v'||export_id
	,'%v'||resupply_id
	from difference_serial
	join purchase on purchase.id = difference_serial.purchase_id
	where purchase.warehouse_id = $1
	;`,
		ToteIDCode,
		ToteIDCode,
		BinIDCode,
		ReceiveIDCode,
		PurchaseIDCode,
		ExportIDCode,
		ResupplyIDCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		ss                                    []Serial
		pickTote, binID, exportID, resupplyID sql.NullString
	)
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
		var s Serial

		err = rows.Scan(
			&s.NanoID,
			&s.ReceiveTote.ID,
			&pickTote,
			&binID,
			&s.Receive.ID,
			&s.Purchase.ID,
			&s.GTIN,
			&exportID,
			&resupplyID,
		)
		if err != nil {
			return nil, err
		}

		if pickTote.Valid {
			s.PickTote.ID = pickTote.String
		}

		if binID.Valid {
			s.Bin.ID = binID.String
		}

		if exportID.Valid {
			s.Export.ID = exportID.String
		}

		if resupplyID.Valid {
			s.Resupply.ID = resupplyID.String
		}

		ss = append(ss, s)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// DifferenceSerialsByPutawayReceive return the serials unsuccessfully putaway of the receiveID
func (db *Data) DifferenceSerialsByPutawayReceive(warehouseID, receiveID string) ([]Serial, error) {
	ds, err := db.DifferenceSerialsByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	// Add the unsuccessfully putaway serials of the receiveID
	var ss []Serial
	for _, s := range ds {
		if s.Receive.ID == receiveID && len(s.Bin.ID) == 0 {
			ss = append(ss, s)
		}
	}

	return ss, nil
}

// DifferenceSerialsByGTINOfPutawayReceive take the result of DifferenceSerialsByPutawayReceive(warehouseID, receiveID string)
// and filter the serials by the gtin parameter
func (db *Data) DifferenceSerialsByGTINOfPutawayReceive(warehouseID, receiveID, gtin string) ([]Serial, error) {
	ds, err := db.DifferenceSerialsByPutawayReceive(warehouseID, receiveID)
	if err != nil {
		return nil, err
	}

	var ss []Serial
	for _, s := range ds {
		if s.GTIN == gtin && len(s.Bin.ID) == 0 {
			ss = append(ss, s)
		}
	}

	return ss, nil
}

func addPackDifferenceSerials(tx *sql.Tx, exportID64 int64, unpackedSerials []Serial) error {
	stmt := `
	insert into difference_serial (activity_id, nanoid, receive_tote, receive_id, purchase_id, gtin) values
	($1
	,$2
	,substring($3 from 5 for 1)::bigint
	,substring($4 from 5 for 1)::bigint
	,substring($5 from 5 for 1)::bigint
	,$6
	)
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, s := range unpackedSerials {
		_, err := tx.ExecContext(ctx, stmt,
			fmt.Sprintf("%v%v", PackIDCode, exportID64),
			s.NanoID,
			s.ReceiveTote.ID,
			s.Receive.ID,
			s.Purchase.ID,
			s.GTIN,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddPackDifferenceSerials adds the picked but unsuccessfully packed serials
func (db *Data) AddPackDifferenceSerials(exportID string) error {
	ss, err := db.UnpackedSerialsByExport(exportID)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	eI, err := id64(exportID, ExportIDCode)
	if err != nil {
		return nil
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = addPackDifferenceSerials(tx, eI, ss)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

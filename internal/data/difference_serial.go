package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (db *Data) AddDifferenceSerials(rc *Receive) error {
	for _, iq := range rc.Items {
		for _, s := range iq.Serials {
			if len(s.Bin.ID) == 0 && len(iq.PutawayNote) != 0 {
				stmt := `
				insert into difference_serial (nanoid, receive_tote, receive_id, purchase_id, gtin) values
				($1
				,substring($2 from 5 for 1)::bigint
				,substring($3 from 5 for 1)::bigint
				,substring($4 from 5 for 1)::bigint
				,$5
				)
				;`

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				_, err := db.DB.ExecContext(ctx, stmt,
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
	--where purchase.warehouse_id = substring($1 from 5 for 1)::bigint
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

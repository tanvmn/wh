package data

import (
	"context"
	"database/sql"
	"time"
)

type PutAwayBin struct {
	Items []ItemQuantity `json:"items,omitempty,omitzero"`
	Bin   `json:"bin,omitempty,omitzero"`
}

func (db *Data) PutawayBins(receiveID string) ([]PutAwayBin, error) {
	rc, err := db.Receive(receiveID)
	if err != nil {
		return nil, err
	}
	// Update actual amount of received units, which is len(Serials) of the each ItemQuantity
	db.AcutalReceiveQuantity(rc)

	bs, err := db.CurrentBinsCapaciy(rc.Purchase.Warehouse.ID)
	if err != nil {
		return nil, err
	}

	m := make(map[string][]ItemQuantity)

	// Calculate the putaway bins
	for _, it := range rc.Items {
		for i := range bs {
			temp, err := db.Item(it.GTIN)
			if err != nil {
				return nil, err
			}
			var iq ItemQuantity

			n := int64(bs[i].Capacity / it.Item.Volume)
			if n >= it.ActualQuantity {
				iq.Item = *temp
				iq.Quantity = it.ActualQuantity
				iq.Serials = it.Serials

				m[bs[i].ID] = append(m[bs[i].ID], iq)
				break
			}
			it.ActualQuantity -= n
			iq.Item = *temp
			iq.Quantity = it.ActualQuantity
			iq.Serials = it.Serials
			m[bs[i].ID] = append(m[bs[i].ID], iq)
		}
	}

	var pbs []PutAwayBin
	for k, v := range m {
		var pb PutAwayBin

		b, err := db.Bin(k)
		if err != nil {
			return nil, err
		}
		pb.Bin = *b
		pb.Items = v

		pbs = append(pbs, pb)
	}

	return pbs, nil
}

func putaway(tx *sql.Tx, rc *Receive) error {
	stmt := `
	update serial
	set bin_id = $1 where nanoid = $2
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, iq := range rc.Items {
		for _, s := range iq.Serials {
			if len(s.Bin.ID) != 0 {
				bI, err := id64(s.Bin.ID, BinIDCode)
				if err != nil {
					return err
				}

				_, err = tx.ExecContext(ctx, stmt, bI, s.NanoID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func setReceiveItemPutawayNote(tx *sql.Tx, rc *Receive) error {
	stmt := `
	update receive_item
	set putaway_note = $1 where purchase_id = $2 and receive_id = $3 and gtin = $4
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, iq := range rc.Items {
		if len(iq.PutawayNote) != 0 {
			pI, err := id64(rc.Purchase.ID, PurchaseIDCode)
			if err != nil {
				return err
			}
			rI, err := id64(rc.ID, ReceiveIDCode)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, stmt, iq.PutawayNote, pI, rI, iq.GTIN)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func setReceivePutawayAt(tx *sql.Tx, receiveID string) error {
	stmt := `
	update receive set putaway_at = now() where id = substring($1 from 5 for 1)::bigint
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, receiveID)
	if err != nil {
		return err
	}

	return nil
}

func setReceivePutawayBy(tx *sql.Tx, rc *Receive) error {
	stmt := `
	update receive set putaway_by = substring($1 from 5 for 1)::bigint where id = substring($2 from 5 for 1)::bigint
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, rc.PutawayAccount.ID, rc.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *Data) Putaway(rc *Receive) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = putaway(tx, rc)
	if err != nil {
		return err
	}

	err = setReceiveItemPutawayNote(tx, rc)
	if err != nil {
		return err
	}

	err = setReceivePutawayBy(tx, rc)
	if err != nil {
		return err
	}

	err = setReceivePutawayAt(tx, rc.ID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// PutawaySerialsByReceive return the serials successfully putaway of a receive
func (db *Data) PutawaySerialsByReceive(rc *Receive) ([]Serial, error) {
	r, err := db.Receive(rc.ID)
	if err != nil {
		return nil, err
	}

	var ss []Serial
	for _, iq := range r.Items {
		for _, s := range iq.Serials {
			if len(s.Bin.ID) != 0 {
				ss = append(ss, s)
			}
		}
	}

	return ss, nil
}

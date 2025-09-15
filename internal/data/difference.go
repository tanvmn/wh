package data

import (
	"fmt"
)

type DifferenceActivity struct {
	ID       string `json:"id,omitempty,omitzero"`
	Name     string `json:"name,omitempty,omitzero"`
	At       string `json:"at,omitempty,omitzero"`
	Result   string `json:"result,omitempty,omitzero"`
	URL      string `json:"url,omitempty,omitzero"`
	*Account `json:"by,omitempty,omitzero"`
}

func (db *Data) DifferenceActivityName(id string) string {
	code := id[:4]
	switch code {
	case ReceiveIDCode:
		return "Nhập hàng"
	case PutawayIDCode:
		return "Cất hàng"
	case PickIDCode:
		return "Lấy hàng"
	case PackIDCode:
		return "Đóng gói"
	default:
		return "none"
	}
}

func (db *Data) doesReceiveProcessHasDifferences(rc *Receive) (bool, error) {
	for _, iq := range rc.Items {
		createdSerials := int64(len(iq.Serials)) // serials created by processing receive
		ss, err := db.DifferenceSerialsByGTINOfPutawayReceive(rc.Purchase.Warehouse.ID, rc.ID, iq.Item.GTIN)
		if err != nil {
			return false, err
		}

		if createdSerials+int64(len(ss)) != iq.Quantity {
			return true, nil
		}
	}

	return false, nil
}

// DifferenceReceiveProcess finds receives that have differences after processing
// and returns the equivalent difference activities
func (db *Data) DifferenceReceiveProcesses(rs []Receive) (as []DifferenceActivity, err error) {
	// filter receives that have difference after processing
	for _, r := range rs {
		diff, err := db.doesReceiveProcessHasDifferences(&r)
		if err != nil {
			return nil, err
		}
		if diff {
			var expectedReceiveQuant, actualReceiveQuant int64
			for _, iq := range r.Items {
				ss, err := db.DifferenceSerialsByGTINOfPutawayReceive(r.Purchase.Warehouse.ID, r.ID, iq.GTIN)
				if err != nil {
					return nil, err
				}

				iq.Serials = append(iq.Serials, ss...)

				expectedReceiveQuant += iq.Quantity
				actualReceiveQuant += int64(len(iq.Serials))
			}

			ac := DifferenceActivity{
				ID:      r.ID,
				Name:    db.DifferenceActivityName(r.ID),
				At:      r.ActualAt,
				Account: &r.ProcessedAccount,
				Result:  fmt.Sprintf("%v / %v", actualReceiveQuant, expectedReceiveQuant),
			}
			as = append(as, ac)
		}
	}

	return as, nil
}

func (db *Data) DifferenceReceivePutaways(rs []Receive) (as []DifferenceActivity, err error) {
	for _, r := range rs {
		ss, err := db.DifferenceSerialsByPutawayReceive(r.Purchase.Warehouse.ID, r.ID)
		if err != nil {
			return nil, err
		}
		if len(ss) != 0 {
			putawaySerials := 0
			for _, iq := range r.Items {
				for _, s := range iq.Serials {
					if len(s.Bin.ID) != 0 && len(s.PickTote.ID) == 0 {
						putawaySerials++
					}
				}
			}

			ac := DifferenceActivity{
				ID:      PutawayIDCode + r.ID[4:],
				Name:    db.DifferenceActivityName(PutawayIDCode + r.ID[4:]),
				At:      r.PutawayAt,
				Account: &r.PutawayAccount,
				Result:  fmt.Sprintf("%v / %v", putawaySerials, putawaySerials+len(ss)),
			}
			as = append(as, ac)
		}
	}

	return as, err
}

func (db *Data) DifferenceActivities(warehouseID string) (as []DifferenceActivity, err error) {
	rs, err := db.Receives(warehouseID)
	if err != nil {
		return nil, err
	}

	// Check for receives of a warehouse that have expected receive quantity different from actual receive quantity
	pcs, err := db.DifferenceReceiveProcesses(rs)
	if err != nil {
		return nil, err
	}
	as = append(as, pcs...)

	// Check for receives of a warehouse that have differences in putaway
	pts, err := db.DifferenceReceivePutaways(rs)
	if err != nil {
		return nil, err
	}
	as = append(as, pts...)

	// Check for export of a warehouse that have differences in picking

	// Check for export of a warehouse that have differences in packing

	return as, nil
}

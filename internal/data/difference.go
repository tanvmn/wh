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

func (db *Data) DifferenceActivities(warehouseID string) (as []any, err error) {
	// Check for receives of a warehouse that have expected receive quantity different from actual receive quantity
	rs, err := db.Receives(warehouseID)
	if err != nil {
		return nil, err
	}

	for _, r := range rs {
		diff, err := db.doesReceiveProcessHasDifferences(&r)
		if err != nil {
			return nil, err
		}
		if diff {
			var expectedReceiveQuant, actualReceiveQuant int64
			for _, iq := range r.Items {
				ss, err := db.DifferenceSerialsByGTINOfPutawayReceive(warehouseID, r.ID, iq.GTIN)
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

			break

		}
	}

	// Check for receives of a warehouse that have differences in putaway

	// Check for export of a warehouse that have differences in picking

	// Check for export of a warehouse that have differences in packing

	return as, nil
}

package data

import (
	"fmt"

	"github.com/tanNguyen2220022/wh/internal/util"
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
	if util.Is01011000(rc.ActualAt) {
		return false, nil
	}
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
// and returns the equivalent difference activities.
// Note that the receives parameters were obtained by a warehouse ID.
func (db *Data) DifferenceReceiveProcesses(warehouseID string) (as []DifferenceActivity, err error) {
	rs, err := db.Receives(warehouseID)
	if err != nil {
		return nil, err
	}

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

// DifferenceReceivePutaways returns equivalent difference activities of receives that have differences after putaway.
// Note that the receives parameters were obtained by a warehouse ID.
func (db *Data) DifferenceReceivePutaways(warehouseID string) (as []DifferenceActivity, err error) {
	rs, err := db.Receives(warehouseID)
	if err != nil {
		return nil, err
	}

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

	return as, nil
}

func (db *Data) DifferenceExportPick(warehouseID string) (as []DifferenceActivity, err error) {
	es, err := db.ExportsByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		if !util.Is01011000(e.PickedAt) {
			for _, iq := range e.Items {
				pickedQuantity := len(iq.Serials)

				if iq.Quantity != int64(pickedQuantity) {
					ddmmyyyy24hmi, err := util.FormatRFC3339(e.PickedAt, util.DDMMYYYY24HMI)
					if err != nil {
						return nil, err
					}

					ac := DifferenceActivity{
						ID:      PickIDCode + e.ID[4:],
						Name:    db.DifferenceActivityName(PickIDCode + e.ID[4:]),
						At:      ddmmyyyy24hmi,
						Account: &e.PickedBy,
						Result:  fmt.Sprintf("%v / %v", pickedQuantity, iq.Quantity),
					}
					as = append(as, ac)
				}
			}
		}
	}
	return as, nil
}

func (db *Data) DifferenceExportPack(warehouseID string) (as []DifferenceActivity, err error) {
	es, err := db.ExportsByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		if !util.Is01011000(e.PackedAt) {
			ps, err := db.Packages(e.ID)
			if err != nil {
				return nil, err
			}

			for _, p := range ps {
				for _, iq := range p.Items {
					if iq.Quantity != int64(len(iq.Serials)) {
						ddmmyyyy24hmi, err := util.FormatRFC3339(e.PickedAt, util.DDMMYYYY24HMI)
						if err != nil {
							return nil, err
						}

						ac := DifferenceActivity{
							ID:      PackIDCode + e.ID[4:],
							Name:    db.DifferenceActivityName(PackIDCode + e.ID[4:]),
							At:      ddmmyyyy24hmi,
							Account: &e.PickedBy,
							Result:  fmt.Sprintf("%v / %v", int64(len(iq.Serials)), iq.Quantity),
						}
						as = append(as, ac)
					}
				}
			}
		}
	}

	return as, nil
}

func (db *Data) DifferenceActivities(warehouseID string) (as []DifferenceActivity, err error) {
	// Check for receives of a warehouse that have expected receive quantity different from actual receive quantity
	processes, err := db.DifferenceReceiveProcesses(warehouseID)
	if err != nil {
		return nil, err
	}
	as = append(as, processes...)

	// Check for receives of a warehouse that have differences in putaway
	putaways, err := db.DifferenceReceivePutaways(warehouseID)
	if err != nil {
		return nil, err
	}
	as = append(as, putaways...)

	// Check for export of a warehouse that have differences in picking
	picks, err := db.DifferenceExportPick(warehouseID)
	if err != nil {
		return nil, err
	}
	as = append(as, picks...)

	// Check for export of a warehouse that have differences in packing
	packs, err := db.DifferenceExportPack(warehouseID)
	if err != nil {
		return nil, err
	}
	as = append(as, packs...)

	return as, nil
}

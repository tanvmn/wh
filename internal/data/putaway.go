package data

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

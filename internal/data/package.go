package data

import gonanoid "github.com/matoous/go-nanoid/v2"

type Package struct {
	NanoID string         `json:"nanoID,omitempty,omitzero"`
	Cap    float64        `json:"cap,omitempty,omitzero"`
	Items  []ItemQuantity `json:"items,omitempty,omitzero"`
	Export `json:"export,omitempty,omitzero"`
}

func (db *Data) NewPackage() (*Package, error) {
	nanoID, err := gonanoid.New()
	if err != nil {
		return nil, err
	}

	return &Package{NanoID: PackageIDCode + nanoID, Cap: 87320}, nil
}

func (db *Data) CalculatedPackages(exportID string) ([]Package, error) {
	iq, err := db.PickedItems(exportID)
	if err != nil {
		return nil, err
	}
	// println("picked items")
	// for _, i := range iq {
	// 	println(i.Item.GTIN, i.Quantity)
	// }
	// println()

	var ps []Package

	// println("calculate packages")
	for i := 0; i < len(iq); {
		p, err := db.NewPackage()
		if err != nil {
			return nil, err
		}

		for p.Cap >= iq[i].Item.Volume {
			// println("current iq:", iq[i].Item.GTIN, "volume:", iq[i].Item.Volume, "quantity:", iq[i].Quantity)
			packedQuant := int64(p.Cap / iq[i].Item.Volume)
			// println("packed quant:", packedQuant)
			if packedQuant >= iq[i].Quantity {
				// println(">=")
				p.Cap -= float64(iq[i].Quantity) * iq[i].Item.Volume
				p.Items = append(p.Items, iq[i])
				// println("package", p.NanoID, iq[i].Item.GTIN, iq[i].Quantity)
				// println()
				if i++; i < len(iq) {
					continue
				} else {
					ps = append(ps, *p)
					return ps, nil
				}
			} else {
				// println("<")
				iq2 := iq[i]
				iq2.Quantity = packedQuant
				p.Items = append(p.Items, iq2)
				// println("package", p.NanoID, iq2.Item.GTIN, iq2.Quantity)
				// println()
				ps = append(ps, *p)
				iq[i].Quantity -= packedQuant
				break
			}
		}
	}

	return ps, nil
}

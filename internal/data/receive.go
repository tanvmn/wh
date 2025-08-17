package data

import (
	"context"
	"time"
)

type Receive struct {
	ID         string       `json:"id,omitempty,omitzero"`
	Purchase   Purchase     `json:"purchase,omitempty,omitzero"`
	Account    Account      `json:"account,omitempty,omitzero"`
	ExpectedAt string       `json:"expectedAt,omitempty,omitzero"`
	ActualAt   string       `json:"actualAt,omitempty,omitzero"`
	CreatedAt  string       `json:"createdAt,omitempty,omitzero"`
	Transfer   Transfer     `json:"transfer,omitempty,omitzero"`
	Items      ItemQuantity `json:"items,omitempty,omitzero"`
}

func (db *Data) ReceiveItemByPurchase(purchaseID string) ([]ItemQuantity, error) {
	i, err := id64(purchaseID, PurchaseIDCode)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	select
	ri.gtin
	,characteristic
	,volume
	,weight
	,brand
	,material
	,color
	,size
	,price
	,currency
	,type
	,shelf_life
	,type||', '||brand||', màu '||color||', cỡ '||size||', '||material
	,sum(quantity)
	from receive_item as ri
	join item on item.gtin = ri.gtin
	where purchase_id = $1
	group by
	ri.gtin
	,characteristic
	,volume
	,weight
	,brand
	,material
	,color
	,size
	,price
	,currency
	,type
	,shelf_life
	;
	`

	rows, err := db.DB.QueryContext(ctx, stmt, i)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			panic(err)
		}
	}()

	var is []ItemQuantity
	for rows.Next() {
		var iq ItemQuantity

		err = rows.Scan(
			&iq.Item.GTIN,
			&iq.Item.Characteristic,
			&iq.Item.Volume,
			&iq.Item.Weight,
			&iq.Item.Brand,
			&iq.Item.Material,
			&iq.Item.Color,
			&iq.Item.Size,
			&iq.Item.Price,
			&iq.Item.Currency,
			&iq.Item.Type,
			&iq.Item.ShelfLife,
			&iq.Item.Name,
			&iq.Quantity,
		)
		is = append(is, iq)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return is, nil
}

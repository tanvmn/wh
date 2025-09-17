package data

import (
	"context"
	"fmt"
	"time"
)

type Resupply struct {
	ID         string         `json:"id,omitempty,omitzero"`
	Status     string         `json:"status,omitempty,omitzero"`
	ExpectedAt string         `json:"expectedAt,omitempty,omitzero"`
	CreatedAt  string         `json:"createdAt,omitempty,omitzero"`
	Items      []ItemQuantity `json:"items,omitempty,omitzero"`
	Account    `json:"account,omitempty,omitzero"`
}

func (db *Data) ResupplyItemsQuantityByWarehouse(warehouseID string) ([]ItemQuantity, error) {
	wI, err := id64(warehouseID, WarehouseIDCode)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(`
	select
	ri.gtin
	,sum(ri.quantity)
	from resupply_item as ri
	join resupply on resupply.id = ri.resupply_id
	join store on store.id = resupply.store_id
	where resupply.status != '%v'
	and store.warehouse_id = $1
	group by ri.gtin
	;`,
		Ended,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	var iqs []ItemQuantity

	for rows.Next() {
		var iq ItemQuantity
		err = rows.Scan(
			&iq.Item.GTIN,
			&iq.Quantity,
		)
		if err != nil {
			return nil, err
		}

		iqs = append(iqs, iq)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return iqs, nil
}

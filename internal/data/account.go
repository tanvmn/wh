package data

type Account struct {
	ID          int64  `json:"id,omitempty,omitzero"`
	BDate       string `json:"bdate,omitempty,omitzero"`
	Name        string `json:"name,omitempty,omitzero"`
	Phone       string `json:"phone,omitempty,omitzero"`
	WarehouseID int64  `json:"warehouseID,omitempty,omitzero"`
	StoreID     int64  `json:"storeID,omitempty,omitzero"`
}

func (d *Data) Account(id int64) (*Account, error) {
	stmt := `select
	id,
	bdate,
	name,
	phone
	from account
	where id=$1`

	var ac Account
	err := d.db.QueryRow(stmt, id).Scan(
		&ac.ID,
		&ac.BDate,
		&ac.Name,
		&ac.Phone,
	)
	if err != nil {
		return nil, err
	}

	return &ac, nil
}

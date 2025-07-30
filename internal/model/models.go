package model

type Account struct {
	ID          int64  `json:"id,omitempty,omitzero"`
	BDate       string `json:"bdate,omitempty,omitzero"`
	Name        string `json:"name,omitempty,omitzero"`
	Phone       string `json:"phone,omitempty,omitzero"`
	WarehouseID int64  `json:"warehouseID,omitempty,omitzero"`
	StoreID     int64  `json:"storeID,omitempty,omitzero"`
}

package data

type Transfer struct {
	ID               string    `json:"id,omitempty,omitzero"`
	ExportWarehouse  Warehouse `json:"exportWarehouse,omitempty,omitzero"`
	ReceiveWarehouse Warehouse `json:"receiveWarehouse,omitempty,omitzero"`
	Account          Account   `json:"account,omitempty,omitzero"`
	ExpectedAt       string    `json:"expectedAt,omitempty,omitzero"`
	CreatedAt        string    `json:"createdAt,omitempty,omitzero"`
}

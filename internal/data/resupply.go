package data

type Resupply struct {
	ID         string  `json:"id,omitempty,omitzero"`
	Account    Account `json:"account,omitempty,omitzero"`
	Store      Store   `json:"store,omitempty,omitzero"`
	ExpectedAt string  `json:"expectedAt,omitempty,omitzero"`
	Status     string  `json:"status,omitempty,omitzero"`
	CreatedAt  string  `json:"createdAt,omitempty,omitzero"`
	Items      []struct {
		Item     Item  `json:"item,omitempty,omitzero"`
		Quantity int64 `json:"quantity,omitempty,omitzero"`
	} `json:"items,omitempty,omitzero"`
}

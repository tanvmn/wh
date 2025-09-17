package data

type Resupply struct {
	ID         string         `json:"id,omitempty,omitzero"`
	Status     string         `json:"status,omitempty,omitzero"`
	ExpectedAt string         `json:"expectedAt,omitempty,omitzero"`
	CreatedAt  string         `json:"createdAt,omitempty,omitzero"`
	Items      []ItemQuantity `json:"items,omitempty,omitzero"`
	Account    `json:"account,omitempty,omitzero"`
}

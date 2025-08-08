package data

type Resupply struct {
	ID         string   `json:"id,omitempty,omitzero"`
	Account    Account `json:"account,omitempty,omitzero"`
	Store      Store   `json:"store,omitempty,omitzero"`
	ExpectedAt string   `json:"expectedAt,omitempty,omitzero"`
	Status     string   `json:"status,omitempty,omitzero"`
	CreatedAt  string   `json:"createdAt,omitempty,omitzero"`
}

type Export struct {
	ID         string    `json:"id,omitempty,omitzero"`
	Resupply   Resupply `json:"resupply,omitempty,omitzero"`
	ExpectedAt string    `json:"expectedAt,omitempty,omitzero"`
	ActualAt   string    `json:"actualAt,omitempty,omitzero"`
	CreatedAt  string    `json:"createdAt,omitempty,omitzero"`
	Transfer   Transfer `json:"transfer,omitempty,omitzero"`
}

package data

type Export struct {
	ID         string   `json:"id,omitempty,omitzero"`
	Resupply   Resupply `json:"resupply,omitempty,omitzero"`
	ExpectedAt string   `json:"expectedAt,omitempty,omitzero"`
	ActualAt   string   `json:"actualAt,omitempty,omitzero"`
	CreatedAt  string   `json:"createdAt,omitempty,omitzero"`
	Transfer   Transfer `json:"transfer,omitempty,omitzero"`
	Items      []struct {
		Item     Item  `json:"item,omitempty,omitzero"`
		Quantity int64 `json:"quantity,omitempty,omitzero"`
	} `json:"items,omitempty,omitzero"`
}
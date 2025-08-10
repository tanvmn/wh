package data

const (
	AwaitingResponse = "Chờ phản hồi"
	AwaitingReceive  = "Chờ nhập"
	Receiving        = "Đang nhập"
	Ended            = "Kết thúc"
	Declined         = "Từ chối"
)

type Purchase struct {
	Warehouse  Warehouse `json:"warehouse,omitempty,omitzero"`
	Account    Account   `json:"account,omitempty,omitzero"`
	ExpectedAt string    `json:"expectedAt,omitempty,omitzero"`
	Status     string    `json:"status,omitempty,omitzero"`
	Supplier   Supplier  `json:"supplier,omitempty,omitzero"`
	ID         string    `json:"purchase,omitempty,omitzero"`
	CreatedAt  string    `json:"createdAt,omitempty,omitzero"`
	Items      []Item    `json:"items,omitempty,omitzero"`
}

type Receive struct {
	ID         string   `json:"id,omitempty,omitzero"`
	Purchase   Purchase `json:"purchase,omitempty,omitzero"`
	Account    Account  `json:"account,omitempty,omitzero"`
	ExpectedAt string   `json:"expectedAt,omitempty,omitzero"`
	ActualAt   string   `json:"actualAt,omitempty,omitzero"`
	CreatedAt  string   `json:"createdAt,omitempty,omitzero"`
	Transfer   Transfer `json:"transfer,omitempty,omitzero"`
}

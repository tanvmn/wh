package data

type Warehouse struct {
	ID      string `json:"id,omitempty,omitzero"`
	Name    string `json:"name,omitempty,omitzero"`
	Address string `json:"address,omitempty,omitzero"`
	Phone   string `json:"phone,omitempty,omitzero"`
	Email   string `json:"email,omitempty,omitzero"`
}

type Bin struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
	Capacity  float32   `json:"capacity,omitempty,omitzero"`
	Shelf     int64     `json:"shelf,omitempty,omitzero"`
	Row       int64     `json:"row,omitempty,omitzero"`
	Col       int64     `json:"col,omitempty,omitzero"`
}

type Tote struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
	Capacity  float32   `json:"capacity,omitempty,omitzero"`
}

type Store struct {
	ID        string    `json:"id,omitempty,omitzero"`
	Name      string    `json:"name,omitempty,omitzero"`
	Address   string    `json:"address,omitempty,omitzero"`
	Phone     string    `json:"phone,omitempty,omitzero"`
	Email     string    `json:"email,omitempty,omitzero"`
	Warehouse Warehouse `json:"warehouse,omitempty,omitzero"`
}

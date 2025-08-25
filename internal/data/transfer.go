package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Transfer struct {
	ID               string    `json:"id,omitempty,omitzero"`
	ExportWarehouse  Warehouse `json:"exportWarehouse,omitempty,omitzero"`
	ReceiveWarehouse Warehouse `json:"receiveWarehouse,omitempty,omitzero"`
	Account          Account   `json:"account,omitempty,omitzero"`
	ExpectedAt       string    `json:"expectedAt,omitempty,omitzero"`
	CreatedAt        string    `json:"createdAt,omitempty,omitzero"`
	Note             string    `json:"note,omitempty,omitzero"`
	Version          int       `json:"version,omitempty,omitzero"`
}

var (
	ErrNoTransfers = errors.New("data: no transfer found")
)

func (db *Data) Transfer(id string) (*Transfer, error) {
	i, err := id64(id, TransferIDCode)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
	select
	'%v'||transfer.id
	,'%v'||export_warehouse
	,'%v'||receive_warehouse
	,'%v'||account_id
	,created_at
	,version
	,note
	,expected_at
	from transfer
	-- join account where account.id = account_id
	where transfer.id = $1
	;`,
		TransferIDCode,
		WarehouseIDCode,
		WarehouseIDCode,
		AccountIDCode,
	)
	var tf Transfer

	err = db.DB.QueryRowContext(ctx, stmt, i).Scan(
		&tf.ID,
		&tf.ExportWarehouse.ID,
		&tf.ReceiveWarehouse.ID,
		&tf.Account.ID,
		&tf.Account.Role,
		&tf.CreatedAt,
		&tf.Version,
		&tf.Note,
		&tf.ExpectedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoTransfers
		}
		return nil, err
	}

	return &tf, nil
}

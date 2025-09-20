package data

import (
	"context"
	"database/sql"
	"time"
)

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

func delExportItemsByResupply(tx *sql.Tx, resupplyID int64) error {
	stmt := `
	delete from export_item where resupply_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, resupplyID)
	return err
}

func delExportByResupply(tx *sql.Tx, resupplyID int64) error {
	stmt := `
	delete from export where resupply_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, resupplyID)
	return err
}

func delExportItems(tx *sql.Tx, exportID int64) error {
	stmt := `
	delete from export_item where export_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, exportID)
	return err
}

func delExport(tx *sql.Tx, exportID int64) error {
	stmt := `
	delete from export where export_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, exportID)
	return err
}

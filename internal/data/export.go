package data

import (
	"context"
	"database/sql"
	"fmt"
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
	if err != nil {
		return err
	}

	return nil
}

func delExportByResupply(tx *sql.Tx, resupplyID int64) error {
	stmt := `
	delete from export where resupply_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, resupplyID)
	if err != nil {
		return err
	}

	return nil
}

func delExportItems(tx *sql.Tx, exportID int64) error {
	stmt := `
	delete from export_item where export_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, exportID)
	if err != nil {
		return err
	}

	return nil
}

func delExport(tx *sql.Tx, exportID int64) error {
	stmt := `
	delete from export where export_id = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, exportID)
	if err != nil {
		return err
	}

	return nil
}

func addExport(tx *sql.Tx, r *Resupply) (id int64, err error) {
	aI, err := id64(r.Account.ID, AccountIDCode)
	if err != nil {
		return 0, err
	}
	rI, err := id64(r.ID, ResupplyIDCode)
	if err != nil {
		return 0, err
	}
	voucherID := fmt.Sprintf("%v%v", VoucherIDCode, rI)

	stmt := `
	insert into export (account_id, expected_at, voucher_id, resupply_id) values
	($1, $2, $3, $4)
	returning id
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tx.QueryRowContext(ctx, stmt, aI, r.ExpectedAt, voucherID, rI).Scan(
		&id,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func addExportItems(tx *sql.Tx, resupplyID, exportID int64, is []ItemQuantity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `
	insert into export_item (export_id, resupply_id, gtin, quantity) values
	($1, $2, $3, $4)
	;`

	for _, i := range is {
		_, err := tx.ExecContext(ctx, stmt, exportID, resupplyID, i.Item.GTIN, i.Quantity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Data) AddExport(resupplyID string) (exportID string, err error) {
	rI, err := id64(resupplyID, ResupplyIDCode)
	if err != nil {
		return "", err
	}

	r, err := db.Resupply(resupplyID)
	if err != nil {
		return "", err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	exportID64, err := addExport(tx, r)
	if err != nil {
		return "", err
	}

	err = addExportItems(tx, rI, exportID64, r.Items)
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v%v", ExportIDCode, exportID64), nil
}

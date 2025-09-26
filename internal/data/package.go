package data

import (
	"context"
	"database/sql"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Package struct {
	NanoID string         `json:"nanoID,omitempty,omitzero"`
	Cap    float64        `json:"cap,omitempty,omitzero"`
	Items  []ItemQuantity `json:"items,omitempty,omitzero"`
	Export `json:"export,omitempty,omitzero"`
}

func (p *Package) NeededPackQuantity() int64 {
	var sum int64
	for _, iq := range p.Items {
		sum += iq.Quantity
	}
	return sum
}

func (p *Package) ActualPackQuantity() int64 {
	var sum int64
	for _, iq := range p.Items {
		sum += int64(len(iq.Serials))
	}
	return sum
}

func (db *Data) NewPackage() (*Package, error) {
	nanoID, err := gonanoid.New()
	if err != nil {
		return nil, err
	}

	return &Package{NanoID: PackageIDCode + nanoID, Cap: 87320}, nil
}

func (db *Data) CalculatedPackages(exportID string) ([]Package, error) {
	iq, err := db.PickedItems(exportID)
	if err != nil {
		return nil, err
	}
	// println("picked items")
	// for _, i := range iq {
	// 	println(i.Item.GTIN, i.Quantity)
	// }
	// println()

	var ps []Package

	// println("calculate packages")
	for i := 0; i < len(iq); {
		p, err := db.NewPackage()
		if err != nil {
			return nil, err
		}

		for p.Cap >= iq[i].Item.Volume {
			// println("current iq:", iq[i].Item.GTIN, "volume:", iq[i].Item.Volume, "quantity:", iq[i].Quantity)
			packedQuant := int64(p.Cap / iq[i].Item.Volume)
			// println("packed quant:", packedQuant)
			if packedQuant >= iq[i].Quantity {
				// println(">=")
				p.Cap -= float64(iq[i].Quantity) * iq[i].Item.Volume
				p.Items = append(p.Items, iq[i])
				// println("package", p.NanoID, iq[i].Item.GTIN, iq[i].Quantity)
				// println()
				if i++; i < len(iq) {
					continue
				} else {
					ps = append(ps, *p)
					return ps, nil
				}
			} else {
				// println("<")
				iq2 := iq[i]
				iq2.Quantity = packedQuant
				p.Items = append(p.Items, iq2)
				// println("package", p.NanoID, iq2.Item.GTIN, iq2.Quantity)
				// println()
				ps = append(ps, *p)
				iq[i].Quantity -= packedQuant
				break
			}
		}
	}

	return ps, nil
}

func addPackages(tx *sql.Tx, packResult *Export) error {
	stmt := `
	insert into package (nanoid, export_id) values
	($1, $2)
	;`

	eI, err := id64(packResult.ID, ExportIDCode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, p := range packResult.Packages {
		_, err := tx.ExecContext(ctx, stmt, p.NanoID, eI)
		if err != nil {
			return err
		}
	}

	return nil
}

func addPackageItems(tx *sql.Tx, packResult *Export) error {
	stmt := `
	insert into package_item (package_nanoid, gtin, quantity, pack_note) values
	($1, $2, $3, $4)
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, p := range packResult.Packages {
		for _, iq := range p.Items {
			_, err := tx.ExecContext(ctx, stmt, p.NanoID, iq.Item.GTIN, iq.Quantity, iq.PackNote)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func addPackageSerials(tx *sql.Tx, packResult *Export) error {
	stmt := `
	insert into package_serial (package_nanoid, gtin, serial_nanoid) values
	($1, $2, $3)
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, p := range packResult.Packages {
		for _, iq := range p.Items {
			for _, s := range iq.Serials {
				_, err := tx.ExecContext(ctx, stmt, p.NanoID, iq.Item.GTIN, s.NanoID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (db *Data) PackageSerialsByPackage(packageNanoID string) ([]Serial, error) {
	stmt := `
	select
	serial_nanoid
	,gtin
	from package_serial
	where package_nanoid = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, packageNanoID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := rows.Close(); err2 != nil {
			panic(err2)
		}
	}()

	var ss []Serial

	for rows.Next() {
		var s Serial

		err = rows.Scan(
			&s.NanoID,
			&s.GTIN,
		)
		if err != nil {
			return nil, err
		}

		ss = append(ss, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ss, err
}

func (db *Data) PackageItems(packageNanoID string) ([]ItemQuantity, error) {
	stmt := `
	select
	gtin
	,quantity
	,pack_note
	from package_item
	where package_nanoid = $1
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, packageNanoID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := rows.Close(); err2 != nil {
			panic(err2)
		}
	}()

	var iqs []ItemQuantity

	for rows.Next() {
		var iq ItemQuantity

		err = rows.Scan(
			&iq.Item.GTIN,
			&iq.Quantity,
			&iq.PackNote,
		)
		if err != nil {
			return nil, err
		}

		i, err := db.Item(iq.Item.GTIN)
		if err != nil {
			return nil, err
		}
		iq.Item = *i

		ss, err := db.PackageSerialsByPackage(packageNanoID)
		if err != nil {
			return nil, err
		}
		for _, s := range ss {
			if s.GTIN == iq.Item.GTIN {
				iq.Serials = append(iq.Serials, s)
			}
		}

		iqs = append(iqs, iq)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return iqs, nil
}

func (db *Data) Packages(exportID string) ([]Package, error) {
	eI, err := id64(exportID, ExportIDCode)
	if err != nil {
		return nil, err
	}

	stmt := `
	select
	nanoid
	,$1||export_id
	from package
	where export_id = $2
	;`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, stmt, ExportIDCode, eI)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := rows.Close(); err2 != nil {
			panic(err2)
		}
	}()

	var ps []Package

	for rows.Next() {
		var p Package

		err = rows.Scan(
			&p.NanoID,
			&p.Export.ID,
		)
		if err != nil {
			return nil, err
		}

		// Add the export to the package
		e, err := db.Export(p.Export.ID)
		if err != nil {
			return nil, err
		}
		p.Export = *e

		// Add the items to the package
		iqs, err := db.PackageItems(p.NanoID)
		if err != nil {
			return nil, err
		}
		p.Items = iqs

		ps = append(ps, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}

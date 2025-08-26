package data

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/tanNguyen2220022/wh/internal/validator"
)

const (
	AccountIDCode   = "ACC-"
	ItemIDCode      = "ITE-"
	SerialIDCode    = "SER-"
	BinIDCode       = "BIN-"
	ToteIDCode      = "TOT-"
	BoxIDCode       = "BOX-"
	StaffIDCode     = "STA-"
	WarehouseIDCode = "WAR-"
	StoreIDCode     = "STO-"
	SupplierIDCode  = "SUP-"
	PurchaseIDCode  = "PUR-"
	ReceiveIDCode   = "REC-"
	ResupplyIDCode  = "RES-"
	ExportIDCode    = "EXP-"
	TransferIDCode  = "TRA-"
)

const (
	AwaitingResponse = "Chờ phản hồi"
	AwaitingReceive  = "Chờ nhập"
	Receiving        = "Đang nhập"
	Ended            = "Kết thúc"
	Declined         = "Từ chối"
)

var (
	ErrInvalidID   = errors.New("data: invalid ID")
	ErrSetConflict = errors.New("data: set conflict")
	ErrCorruptedData = errors.New("data: corrupted data")
)

type Data struct {
	DB     *sql.DB
	logger *slog.Logger
}

func NewData(db *sql.DB, lg *slog.Logger) *Data {
	return &Data{
		DB:     db,
		logger: lg,
	}
}

// Range returns the string of sql range syntax in a where clause
func Range(n int64) string {
	var r string

	for i := range n {
		if r == "" {
			r += "$" + strconv.FormatInt(i+1, 10)
		} else {
			r += ",$" + strconv.FormatInt(i+1, 10)
		}
	}
	r = "(" + r + ")"

	return r
}

// id64 checks if id is at least 5 chars and if the code part is one of permittedCodes,
// then parses the number part to an int64
func id64(id string, permittedCodes string) (int64, error) {
	va := validator.Validator{}

	s := []string{permittedCodes}
	va.Check(
		validator.MinChars(id, 5) && validator.Permitted(id[:4], s...),
		fmt.Sprintf("ID %v is less than 5 chars or the code is not within %v", id, permittedCodes),
	)
	if !va.Valid() {
		return 0, fmt.Errorf("%w: %v", ErrInvalidID, va.Message())
	}

	i, err := strconv.ParseInt(id[4:], 10, 64)
	if err != nil || i < 1 {
		return 0, fmt.Errorf("%w: ID %v, must be a number >= 1 after ID code", ErrInvalidID, id)
	}

	return i, nil
}

package main

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
	"github.com/tanNguyen2220022/wh/internal/validator"
)

var (
	ErrInvalidPurchase = errors.New("invalid purchase")
)

// validatePurchaseAdd validates *data.Purchase against the current data
// and set ExpectedAt to datetime if the received input is valid
func (ap *application) validatePurchaseAdd(pc *data.Purchase) error {
	var err error
	va := validator.Validator{}

	// Validate chosen date time
	dt, err := util.FormatDateTTime(pc.ExpectedAt, time.DateTime)
	if err != nil {
		va.AddErr(fmt.Sprintf("%v: %v", err, pc.ExpectedAt))
	}
	pc.ExpectedAt = dt

	// Validate warehouse's existence
	wh, err := ap.data.Warehouse(pc.Warehouse.ID)
	if errors.Is(err, data.ErrNoWarehouses) {
		va.AddErr(err.Error())
	} else if err != nil {
		ap.logger.Error(err.Error())
		return err
	}

	// Validate account's existence
	ac, err := ap.data.Account(pc.Account.ID)
	if err != nil {
		ap.logger.Error(err.Error())
		return err
	}

	// If both account and warehouse exist
	if ac != nil && wh != nil {
		// Validate if the account is from the warehouse when account's role isn't Admin and isn't HeadAccount
		if ac.Role != data.Admin && ac.Role != data.HeadAccountant {
			from, err := ap.data.IsAccountFromWarehouse(ac.ID, wh.ID)
			if err != nil {
				ap.logger.Error(err.Error())
				return err
			}
			va.Check(from, fmt.Sprintf("Account %v isn't from warehouse %v, yet the account still made the purchase", pc.Account.ID, pc.Warehouse.ID))
		}
	}

	// Validate supplier's existence
	sp, err := ap.data.Supplier(pc.Supplier.ID)
	if errors.Is(err, data.ErrNoSuppliers) {
		va.AddErr(err.Error())
	} else if err != nil {
		ap.logger.Error(err.Error())
		return err
	}

	// Validate chosen items' existence when the supplier exists
	if len(pc.Items) == 0 {
		va.AddErr("No items in purchase")
	} else if sp != nil {
		gtins, err := ap.data.GTINsBySupplier(pc.Supplier.ID)
		if errors.Is(err, data.ErrInvalidID) {
			va.AddErr(err.Error())
		} else if err != nil {
			ap.logger.Error(err.Error())
			return err
		}
		va.Check(len(gtins) != 0, fmt.Sprintf("Chosen supplier %v doesn't supply any items, yet there are still %v item(s) in purchase", pc.Supplier.ID, len(pc.Items)))

		for _, i := range pc.Items {
			// Validate item's existence
			it, err := ap.data.Item(i.Item.GTIN)
			if errors.Is(err, data.ErrNoItems) {
				va.AddErr(err.Error())
			} else if err != nil {
				ap.logger.Error(err.Error())
				return err
			}

			if it != nil {
				if len(gtins) != 0 {
					va.Check(slices.Contains(gtins, i.Item.GTIN), fmt.Sprintf("GTIN %v isn't supplied by supplier %v, yet it's still in purchase", i.Item.GTIN, pc.Supplier.ID))
				}
				va.Check(i.Quantity > 0, fmt.Sprintf("GTIN %v, quantity must be > 0", i.Item.GTIN))
			}
		}
	}

	if !va.Valid() {
		return fmt.Errorf("%w: %v", ErrInvalidPurchase, va.Message())
	}

	return nil
}

func (ap *application) addPurchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pc data.Purchase

		err := ap.decodeJSON(w, r, &pc)
		if err != nil {
			ap.logger.Error(util.ErrLine)

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Get account ID in context for validating purchase before adding
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(fmt.Errorf("%w: %v", ErrConvertCtxVal, "cannot convert context accountID to string").Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pc.Account.ID = aID

		// Validate the purchase
		err = ap.validatePurchaseAdd(&pc)
		if errors.Is(err, ErrInvalidPurchase) {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check against warehouse capacity
		enough, err := ap.data.CheckCapacity(pc.Items, pc.Warehouse.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !enough {
			ap.logger.Error("Not enough capacity")
			http.Error(w, fmt.Sprintf("Kho %v hiện không đủ sức chứa", pc.Warehouse.ID), http.StatusUnprocessableEntity)
			return
		}

		id, _, err := ap.data.AddPurchase(&pc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Đã thêm yêu cầu nhập ID: %v", id)
	})
}

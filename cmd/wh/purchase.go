package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/util"
	"github.com/tanvmn/wh/internal/validator"
)

var (
	ErrInvalidPurchase      = errors.New("invalid purchase")
	ErrInvalidPurchaseItems = errors.New("invalid purchase items")
)

// validatePurchaseAdd validates *data.Purchase against the current data
// and set ExpectedAt to datetime if the received input is valid
func (a *app) validatePurchaseAdd(pc *data.Purchase) error {
	var err error
	va := validator.Validator{}

	// Validate chosen date time
	_, err = util.FormatDateTTime(pc.ExpectedAt, util.DateTTime)
	if err != nil {
		va.AddErr(err.Error())
	}

	// Validate warehouse's existence
	wh, err := a.data.Warehouse(pc.Warehouse.ID)
	if errors.Is(err, data.ErrNoWarehouses) {
		va.AddErr(err.Error())
	} else if err != nil {
		a.log.Error(err.Error())
		return err
	}

	// Validate account's existence
	ac, err := a.data.Account(pc.Account.ID)
	if err != nil {
		a.log.Error(err.Error())
		return err
	}

	// If both account and warehouse exist
	if ac != nil && wh != nil {
		// Validate if the account is from the warehouse when account's role isn't Admin and isn't HeadAccount
		if ac.Role != data.Admin && ac.Role != data.HeadAccountant {
			from, err := a.data.IsAccountFromWarehouse(ac.ID, wh.ID)
			if err != nil {
				a.log.Error(err.Error())
				return err
			}
			va.Check(from, fmt.Sprintf("Account %v isn't from warehouse %v, yet the account still made the purchase", pc.Account.ID, pc.Warehouse.ID))
		}
	}

	// Validate supplier's existence
	sp, err := a.data.Supplier(pc.Supplier.ID)
	if errors.Is(err, data.ErrNoSuppliers) {
		va.AddErr(err.Error())
	} else if err != nil {
		a.log.Error(err.Error())
		return err
	}

	// Validate chosen items' existence when the supplier exists
	if len(pc.Items) == 0 {
		va.AddErr("No items in purchase")
	} else if sp != nil {
		for _, i := range pc.Items {
			// Validate item's existence
			it, err := a.data.Item(i.Item.GTIN)
			if errors.Is(err, data.ErrNoItems) {
				va.AddErr(err.Error())
			} else if err != nil {
				a.log.Error(err.Error())
				return err
			}

			if it != nil {
				from, err := a.data.IsGTINBySupplier(i.Item.GTIN, pc.Supplier.ID)
				if err != nil {
					a.log.Error(err.Error())
					return err
				}
				va.Check(from, fmt.Sprintf("GTIN %v isn't supplied by supplier %v, yet it's still in purchase", i.Item.GTIN, pc.Supplier.ID))
				va.Check(i.Quantity > 0, fmt.Sprintf("GTIN %v, quantity must be > 0", i.Item.GTIN))
			}
		}
	}

	if !va.Valid() {
		return fmt.Errorf("%w\n%v", ErrInvalidPurchase, va.Message())
	}

	return nil
}

func (a *app) addPurchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pc data.Purchase

		err := a.decodeJSON(w, r, &pc)
		if err != nil {
			a.log.Error(util.ErrLine)

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
			a.log.Error(fmt.Errorf("%w: %v", ErrConvertCtxVal, "cannot convert context accountID to string").Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pc.Account.ID = aID

		// Validate the purchase
		err = a.validatePurchaseAdd(&pc)
		if errors.Is(err, ErrInvalidPurchase) {
			a.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check against warehouse capacity
		enough, err := a.data.CheckCapacity(pc.Items, pc.Warehouse.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !enough {
			a.log.Error("Not enough capacity")
			http.Error(w, fmt.Sprintf("Kho %v hiện không đủ sức chứa", pc.Warehouse.ID), http.StatusUnprocessableEntity)
			return
		}

		// Add the purchase
		id, _, err := a.data.AddPurchase(&pc)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Get the newly added purchase to supply data for template
		p, err := a.data.Purchase(id)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.ExpectedAt, err = util.FormatDateTTime(p.ExpectedAt, util.DDMMYYYY24HMI)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Send purchase email to supplier in the background with a new go routine
		// ap.background(func() {
		// 	err = ap.mailer.Send(p.Supplier.Email, "purchase_mail", p)
		// 	if err != nil {
		// 		ap.logger.Error(err.Error())
		// 		return
		// 	}
		// })

		// w.WriteHeader(http.StatusCreated)
		// fmt.Fprintf(w, "Đã thêm yêu cầu nhập ID: %v", id)
		http.Redirect(w, r, fmt.Sprintf("%v/purchase/%v", domain, id), http.StatusSeeOther)
	})
}

func (a *app) purchasePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		pc, err := a.data.Purchase(id)
		if err != nil {
			a.log.Error(err.Error())

			if errors.Is(err, data.ErrNoPurchases) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu nhập ID: %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		data, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if pc != nil {
			data.Purchase = *pc

			is, err := a.data.ItemsBySupplier(pc.Supplier.ID)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			for _, i := range is {
				picked := false

				for _, pi := range pc.Items {
					if i.GTIN == pi.Item.GTIN {
						picked = true
						break
					}
				}
				if picked {
					picked = false
				} else {
					data.Items = append(data.Items, i)
				}
			}
		}

		if err := a.render(w, http.StatusOK, "purchase", data); err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) addPurchasePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := a.render(w, http.StatusOK, "purchase_add", data); err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) validatePurchaseSet(pc *data.Purchase) error {
	pur, err := a.data.Purchase(pc.ID)
	if errors.Is(err, data.ErrNoPurchases) {
		return fmt.Errorf("%w:\n%v: ID %v", ErrInvalidPurchase, err.Error(), pc.ID)
	} else if err != nil {
		a.log.Error(err.Error())
		return err
	}

	if !(pc.Warehouse.ID == pur.Warehouse.ID && pc.Supplier.ID == pur.Supplier.ID) {
		return fmt.Errorf("%w:\n%v: ID %v", ErrInvalidPurchase, "Kho nhận và nhà cung cấp của yêu cầu nhập ID: %v không thể thay đổi", pc.ID)
	}

	va := validator.Validator{}

	// Check if the new ExpectedAt is the same or after the old one
	newT, err := time.Parse(util.DateTTime, pc.ExpectedAt)
	if err != nil {
		a.log.Error(err.Error())
		return err
	}
	oldT, err := time.Parse(util.DateTTime, pur.ExpectedAt)
	if err != nil {
		a.log.Error(err.Error())
		return err
	}
	va.Check(oldT.Compare(newT) <= 0, fmt.Sprintf("Thời điểm muốn nhận: mới %v không thể trước cũ %v", pc.ExpectedAt, pur.ExpectedAt))

	// Check if the status of the old purchase is still awaiting response or awaiting receive
	va.Check(pur.Status == data.AwaitingResponse || pur.Status == data.AwaitingReceive, fmt.Sprintf("Yêu cầu nhập ID %v đã có ít nhất 1 phiếu nhập được xử lý", pc.ID))

	err = a.validatePurchaseAdd(pc)
	if errors.Is(err, ErrInvalidPurchase) {
		va.AddErr(err.Error())
	} else if err != nil {
		a.log.Error(err.Error())
		return err
	}

	err = a.validatePurchaseItemsSet(pc)
	if errors.Is(err, ErrInvalidPurchaseItems) {
		va.AddErr(err.Error())
	} else if err != nil {
		a.log.Error(err.Error())
		return err
	}

	if !va.Valid() {
		return fmt.Errorf("%w:\n%v", ErrInvalidPurchase, va.Message())
	}

	return nil
}

func (a *app) validatePurchaseItemsSet(pc *data.Purchase) error {
	is, err := a.data.ReceiveItemsByPurchase(pc.ID)
	if err != nil {
		a.log.Error(err.Error())
		return err
	}

	va := validator.Validator{}

	if len(is) != 0 {
		for _, rI := range is {
			contained := false

			for _, pI := range pc.Items {
				if rI.Item.GTIN == pI.Item.GTIN {
					contained = true
					va.Check(rI.Quantity <= pI.Quantity, fmt.Sprintf("GTIN %v, số lượng: %v < %v của tổng các phiếu nhập", pI.Item.GTIN, pI.Quantity, rI.Quantity))
				}
			}

			if contained {
				contained = false
				continue
			}
			va.AddErr(fmt.Sprintf("GTIN %v có trong ít nhất 1 phiếu nhập nhưng KHÔNG có trong yêu cầu nhập", rI.Item.GTIN))
		}
	}

	if !va.Valid() {
		return fmt.Errorf("%w:\n%v", ErrInvalidPurchaseItems, va.Message())
	}

	return nil
}

func (a *app) setPurchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pc data.Purchase

		err := a.decodeJSON(w, r, &pc)
		if err != nil {
			a.log.Error(util.ErrLine)

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Get account ID in context for validating purchase before setting
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			a.log.Error(fmt.Errorf("%w: %v", ErrConvertCtxVal, "cannot convert context accountID to string").Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pc.Account.ID = aID

		// Begin validating purchase before setting
		err = a.validatePurchaseSet(&pc)
		if errors.Is(err, ErrInvalidPurchase) {
			a.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			a.log.Error(util.ErrLine)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Setting the purchase
		err = a.data.SetPurchase(&pc)
		if errors.Is(err, data.ErrSetConflict) {
			a.log.Error(err.Error())
			http.Error(w, fmt.Sprintf("Yêu cầu nhập ID: %v có thể đã được sửa bởi tài khoản khác khi bạn chưa hoàn thành.\nHãy tải lại và thực hiện lại", pc.ID), http.StatusUnprocessableEntity)
			return
		} else if err != nil {
			a.log.Error(err.Error())
			http.Error(w, err.Error(), 520)
			return
		}

		// Get the set purchase to supply data for template
		p, err := a.data.Purchase(pc.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.ExpectedAt, err = util.FormatDateTTime(p.ExpectedAt, util.DDMMYYYY24HMI)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Send purchase email to supplier in the background with a new go routine
		// ap.background(func() {
		// 	err = ap.mailer.Send(p.Supplier.Email, "purchase_mail", p)
		// 	if err != nil {
		// 		ap.logger.Error(err.Error())
		// 		return
		// 	}
		// })

		http.Redirect(w, r, fmt.Sprintf("%v/purchase/%v", domain, pc.ID), http.StatusSeeOther)
	})
}

func (a *app) purchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, err := a.data.Purchase(r.PathValue("id"))
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.writeJSON(w, http.StatusOK, pc, nil)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) delPurchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		pc, err := a.data.Purchase(id)
		if err != nil {
			a.log.Error(err.Error())

			if errors.Is(err, data.ErrNoPurchases) || errors.Is(err, data.ErrInvalidID) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu nhập ID: %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		if !(pc.Status == data.AwaitingResponse || pc.Status == data.AwaitingReceive) {
			a.log.Error(fmt.Sprintf("%v; ID: %v", data.ErrPurchaseReceived, id))
			http.Error(w, fmt.Sprintf("Yêu cầu nhập ID: %v đã nhập ít nhất 1 lần", id), http.StatusBadRequest)
			return
		}

		err = a.data.DelPurchase(id)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		pc.ExpectedAt, err = util.FormatDateTTime(pc.ExpectedAt, util.DDMMYYYY24HMI)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Send purchase email to supplier in the background with a new go routine
		// ap.background(func() {
		// 	err = ap.mailer.Send(pc.Supplier.Email, "purchase_del_mail", pc)
		// 	if err != nil {
		// 		ap.logger.Error(err.Error())
		// 		return
		// 	}
		// })

		http.Redirect(w, r, "/purchases", http.StatusSeeOther)
	})
}

func (a *app) purchasesPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v; %v", ErrConvertCtxVal, authenticatedCtxWarehouseID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ps, err := a.data.Purchases(wID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		data.Purchases = ps

		err = a.render(w, http.StatusOK, "purchases", data)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
	"github.com/tanNguyen2220022/wh/internal/validator"
)

var (
	ErrInvalidPurchase      = errors.New("invalid purchase")
	ErrInvalidPurchaseItems = errors.New("invalid purchase items")
)

// validatePurchaseAdd validates *data.Purchase against the current data
// and set ExpectedAt to datetime if the received input is valid
func (ap *application) validatePurchaseAdd(pc *data.Purchase) error {
	var err error
	va := validator.Validator{}

	// Validate chosen date time
	_, err = util.FormatDateTTime(pc.ExpectedAt, util.DateTTime)
	if err != nil {
		va.AddErr(err.Error())
	}

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
				from, err := ap.data.IsGTINBySupplier(i.Item.GTIN, pc.Supplier.ID)
				if err != nil {
					ap.logger.Error(err.Error())
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

		// Add the purchase
		id, _, err := ap.data.AddPurchase(&pc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Get the newly added purchase to supply data for template
		p, err := ap.data.Purchase(id)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.ExpectedAt, err = util.FormatDateTTime(p.ExpectedAt, util.DDMMYYYY24HMI)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Send purchase email to supplier in the background with a new go routine
		ap.background(func() {
			err = ap.mailer.Send(p.Supplier.Email, "purchase_mail", p)
			if err != nil {
				ap.logger.Error(err.Error())
				return
			}
		})

		// w.WriteHeader(http.StatusCreated)
		// fmt.Fprintf(w, "Đã thêm yêu cầu nhập ID: %v", id)
		http.Redirect(w, r, fmt.Sprintf("%v/purchase/%v", domain, id), http.StatusSeeOther)
	})
}

func (ap *application) purchasePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		pc, err := ap.data.Purchase(id)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoPurchases) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu nhập ID: %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		data, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if pc != nil {
			data.Purchase = *pc

			is, err := ap.data.ItemsBySupplier(pc.Supplier.ID)
			if err != nil {
				ap.logger.Error(err.Error())
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

		if err := ap.render(w, http.StatusOK, "purchase_set", data); err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) addPurchasePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := ap.render(w, http.StatusOK, "purchase_add", data); err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) validatePurchaseSet(pc *data.Purchase) error {
	pur, err := ap.data.Purchase(pc.ID)
	if errors.Is(err, data.ErrNoPurchases) {
		return fmt.Errorf("%w:\n%v: ID %v", ErrInvalidPurchase, err.Error(), pc.ID)
	} else if err != nil {
		ap.logger.Error(err.Error())
		return err
	}

	if !(pc.Warehouse.ID == pur.Warehouse.ID && pc.Supplier.ID == pur.Supplier.ID) {
		return fmt.Errorf("%w:\n%v: ID %v", ErrInvalidPurchase, "Kho nhận và nhà cung cấp của yêu cầu nhập ID: %v không thể thay đổi", pc.ID)
	}

	va := validator.Validator{}

	// Check if the new ExpectedAt is the same or after the old one
	newT, err := time.Parse(util.DateTTime, pc.ExpectedAt)
	if err != nil {
		ap.logger.Error(err.Error())
		return err
	}
	oldT, err := time.Parse(util.DateTTime, pur.ExpectedAt)
	if err != nil {
		ap.logger.Error(err.Error())
		return err
	}
	va.Check(oldT.Compare(newT) <= 0, fmt.Sprintf("Thời điểm muốn nhận: mới %v không thể trước cũ %v", pc.ExpectedAt, pur.ExpectedAt))

	// Check if the status of the old purchase is still awaiting response or awaiting receive
	va.Check(pur.Status == data.AwaitingResponse || pur.Status == data.AwaitingReceive, fmt.Sprintf("Yêu cầu nhập ID %v đã có ít nhất 1 phiếu nhập được xử lý", pc.ID))

	err = ap.validatePurchaseAdd(pc)
	if errors.Is(err, ErrInvalidPurchase) {
		va.AddErr(err.Error())
	} else if err != nil {
		ap.logger.Error(err.Error())
		return err
	}

	err = ap.validatePurchaseItemsSet(pc)
	if errors.Is(err, ErrInvalidPurchaseItems) {
		va.AddErr(err.Error())
	} else if err != nil {
		ap.logger.Error(err.Error())
		return err
	}

	if !va.Valid() {
		return fmt.Errorf("%w:\n%v", ErrInvalidPurchase, va.Message())
	}

	return nil
}

func (ap *application) validatePurchaseItemsSet(pc *data.Purchase) error {
	is, err := ap.data.ReceiveItemsByPurchase(pc.ID)
	if err != nil {
		ap.logger.Error(err.Error())
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

func (ap *application) setPurchase() http.Handler {
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

		// Get account ID in context for validating purchase before setting
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(fmt.Errorf("%w: %v", ErrConvertCtxVal, "cannot convert context accountID to string").Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pc.Account.ID = aID

		// Begin validating purchase before setting
		err = ap.validatePurchaseSet(&pc)
		if errors.Is(err, ErrInvalidPurchase) {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			ap.logger.Error(util.ErrLine)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Setting the purchase
		err = ap.data.SetPurchase(&pc)
		if errors.Is(err, data.ErrSetConflict) {
			ap.logger.Error(err.Error())
			http.Error(w, fmt.Sprintf("Yêu cầu nhập ID: %v có thể đã được sửa bởi tài khoản khác khi bạn chưa hoàn thành.\nHãy tải lại và thực hiện lại", pc.ID), http.StatusUnprocessableEntity)
			return
		} else if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), 520)
			return
		}

		// Get the set purchase to supply data for template
		p, err := ap.data.Purchase(pc.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.ExpectedAt, err = util.FormatDateTTime(p.ExpectedAt, util.DDMMYYYY24HMI)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Send purchase email to supplier in the background with a new go routine
		ap.background(func() {
			err = ap.mailer.Send(p.Supplier.Email, "purchase_mail", p)
			if err != nil {
				ap.logger.Error(err.Error())
				return
			}
		})

		http.Redirect(w, r, fmt.Sprintf("%v/purchase/%v", domain, pc.ID), http.StatusSeeOther)
	})
}

func (ap *application) purchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, err := ap.data.Purchase(r.PathValue("id"))
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, pc, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) delPurchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		pc, err := ap.data.Purchase(id)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoPurchases) || errors.Is(err, data.ErrInvalidID) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu nhập ID: %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		if !(pc.Status == data.AwaitingResponse || pc.Status == data.AwaitingReceive) {
			ap.logger.Error(fmt.Sprintf("%v; ID: %v", data.ErrPurchaseReceived, id))
			http.Error(w, fmt.Sprintf("Yêu cầu nhập ID: %v đã nhập ít nhất 1 lần", id), http.StatusBadRequest)
			return
		}

		err = ap.data.DelPurchase(id)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		pc.ExpectedAt, err = util.FormatDateTTime(pc.ExpectedAt, util.DDMMYYYY24HMI)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Send purchase email to supplier in the background with a new go routine
		ap.background(func() {
			err = ap.mailer.Send(pc.Supplier.Email, "purchase_del_mail", pc)
			if err != nil {
				ap.logger.Error(err.Error())
				return
			}
		})

		http.Redirect(w, r, "/purchases", http.StatusSeeOther)
	})
}

func (ap *application) purchasesPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps, err := ap.data.Purchases()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		data.Purchases = ps

		err = ap.render(w, http.StatusOK, "purchases", data)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tanNguyen2220022/wh/internal/data"
)

// unreceivePurchaseItems returns quantities of items in purchase that haven't been added to any receives
func (ap *application) unreceivedPurchaseItems(pc *data.Purchase) ([]data.ItemQuantity, error) {
	ris, err := ap.data.ReceiveItemsByPurchase(pc.ID)
	if err != nil {
		return nil, err
	}
	if ris == nil {
		return pc.Items, nil
	}

	var iqs []data.ItemQuantity
	copy(pc.Items, iqs)

	pi := pc.Items
	for i := range pi {
		for _, ri := range ris {
			if pi[i].Item.ID == ri.Item.ID {
				if pi[i].Quantity < ri.Quantity {
					err = fmt.Errorf("purchase item %v's quantity is less then added to receives", pi[i].Item.ID)
					ap.logger.Error(err.Error())
					return nil, err
				}

				iq := pi[i]
				iq.Quantity = pi[i].Quantity - ri.Quantity
				iqs = append(iqs, iq)
			}
		}
	}

	return iqs, nil
}

func (ap *application) addReceivePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the purchase ID
		pID := r.URL.Query().Get("purchase")

		// get the purchase data
		pc, err := ap.data.Purchase(pID)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoPurchases) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu nhập ID: %v", pID), http.StatusBadRequest)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v; %v", ErrConvertCtxVal.Error(), aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// if purchase's receive add ID is not ACC-0 or the authenticated ID, then "another acc is adding receive to this pur, please wait and retry later"
		if pc.ReceiveAddOwner != data.AccountIDCode+"0" && pc.ReceiveAddOwner != aID {
			s := fmt.Sprintf("Một tài khoản khác đang thêm phiếu nhập cho yêu cầu nhập ID %v.\nHãy thử lại sau", pc.ID)
			ap.logger.Error(s + "; owner" + pc.ID)
			http.Error(w, s, http.StatusUnprocessableEntity)
			return
		}

		// if receive_add_owner is 0 then claim receive_add_owner
		if pc.ReceiveAddOwner == data.AccountIDCode+"0" {
			err = ap.data.ClaimReceiveAddOwner(pc.ID, aID)
			if err != nil {
				ap.logger.Error(err.Error())

				if errors.Is(err, data.ErrAddReceiveConflict) {
					http.Error(w, fmt.Sprintf("Yêu cầu nhập ID %v có thể đang được thêm phiếu bởi 1 tài khoản khác.\nHãy thử lại sau", pc.ID), http.StatusUnprocessableEntity)
				} else {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				return
			}
			ap.background(func() {
				time.Sleep(20 * time.Minute)
				println("begin unclaiming receive add owner", aID)

				err2 := ap.data.UnclaimReceiveAddOwner(pc.ID, aID)
				if err2 != nil {
					ap.logger.Error(err2.Error())
					panic(err)
				}
			})
		}

		// else serve the add receive page response
		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Purchase = *pc

		iqs, err := ap.unreceivedPurchaseItems(pc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.ItemQuantitys = iqs

		err = ap.render(w, http.StatusOK, "receive_add", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

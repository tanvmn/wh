package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
	"github.com/tanNguyen2220022/wh/internal/validator"
)

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

		// if all items of purchase are added to receives then response the client and return
		upi, err := ap.data.UnreceivedPurchaseItems(pc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(upi) == 0 {
			s := fmt.Sprintf("Tất cả hàng của yêu cầu nhập ID %v đã được thêm vào các phiếu nhập", pc.ID)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusUnprocessableEntity)
			return
		}

		// if purchase's receive add ID is not ACC-0 or the authenticated ID, then "another acc is adding receive to this pur, please wait and retry later"
		if pc.ReceiveAddOwner != data.AccountIDCode+"0" && pc.ReceiveAddOwner != aID {
			s := fmt.Sprintf("Một tài khoản khác đang thêm phiếu nhập cho yêu cầu nhập ID %v.\nHãy thử lại sau", pc.ID)
			ap.logger.Error(s + "; add_receive_owner " + pc.ID)
			http.Error(w, s, http.StatusUnprocessableEntity)
			return
		} else if pc.ReceiveAddOwner == data.AccountIDCode+"0" { // if receive_add_owner is 0 then claim receive_add_owner
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
			// ap.background(func() {
			// 	fmt.Print("\nREMEMBER, a new goroutine is about to unclaim the add receive owner in the background\n\n")
			// 	time.Sleep(7 * time.Minute)
			// 	// time.Sleep(4 * time.Second)
			// 	println("begin unclaiming receive add owner", aID)

			// 	err2 := ap.data.UnclaimReceiveAddOwner(pc.ID, aID)
			// 	if err2 != nil {
			// 		ap.logger.Error(err2.Error())
			// 		panic(err)
			// 	}

			// 	println("finished unclaiming receive add owner", aID)
			// })
		}

		// else serve the add receive page if there items of purchase that are not added to receive
		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Purchase = *pc
		td.ItemQuantitys = upi

		err = ap.render(w, http.StatusOK, "receive_add", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) validateReceive(rc *data.Receive) error {
	pc, err := ap.data.Purchase(rc.Purchase.ID)
	if err != nil {
		return err
	}

	va := validator.Validator{}
	for _, ri := range rc.Items {
		for _, pi := range pc.Items {
			if ri.Item.GTIN == pi.Item.GTIN {
				va.Check(ri.Quantity <= pi.Quantity, fmt.Sprintf("Receive item %v, quantity %v > purchase item %v, quantity %v in purchase %v", ri.Item.GTIN, ri.Quantity, pi.Item.GTIN, pi.Quantity, pc.ID))
				break
			}
		}
	}

	if !va.Valid() {
		return errors.New(va.Message())
	}
	return nil
}

func (ap *application) addReceive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			rc  data.Receive
			err error
		)

		err = ap.decodeJSON(w, r, &rc)
		if err != nil {
			ap.logger.Error(err.Error())

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		err = ap.validateReceive(&rc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if the account is eligible to add receive for purchase
		pc, err := ap.data.Purchase(rc.Purchase.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(ErrConvertCtxVal.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if pc.ReceiveAddOwner == data.AccountIDCode+"0" {
			s := fmt.Sprintf("Đã hết hạn 7ph để tạo phiếu cho yêu cầu nhập ID %v.\nHãy tải lại trang và thực hiện lại", pc.ID)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusUnprocessableEntity)
			return
		} else if pc.ReceiveAddOwner != aID {
			ap.logger.Error(fmt.Sprintf("Account %v received the add receive page and made an add receive request to server, yet the current add receive owner is %v", aID, pc.ReceiveAddOwner))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check if purchase still has items that have not been added to receives.
		// If there aren't, but this request is still present then there has to be an logic error somewhere
		uri, err := ap.data.UnreceivedPurchaseItems(pc)
		if err != nil {
			ap.logger.Error(ErrConvertCtxVal.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(uri) == 0 {
			ap.logger.Error("All items of purchase %v are added to receives, but somehow add receive request (POST) is still made")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Start to add the receive
		// Also unclaims receive add owner at this step
		rc.Account.ID = aID
		err = ap.data.AddReceive(&rc)
		if err != nil {
			ap.logger.Error(err.Error())

			var pgErr *pq.Error
			if errors.As(err, &pgErr) {
				http.Error(w, pgErr.Message, http.StatusBadRequest)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Update purchase status to data.AwaitingReceive if the current status is data.AwaintingResponse
		if pc.Status == data.AwaitingResponse {
			err = ap.data.UpdatePurchaseStatus(pc.ID, pc.Status, data.AwaitingReceive)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else if pc.Status != data.AwaitingReceive && pc.Status != data.Receiving {
			ap.logger.Error(fmt.Sprintf("Purchase %v's current status is %v, but there is a request to add receive made to it", pc.ID, pc.Status))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/receive/"+rc.ID, http.StatusSeeOther)
	})
}

func (ap *application) receivePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		rc, err := ap.data.Receive(id)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập ID: %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Receive = *rc

		td.ItemQuantitys, err = ap.data.UnreceivedPurchaseItemsOpt(rc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.render(w, http.StatusOK, "receive", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) setReceive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rc data.Receive

		err := ap.decodeJSON(w, r, &rc)
		if err != nil {
			ap.logger.Error(err.Error())

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		rcp, err := ap.data.Receive(rc.ID)
		if err != nil {
			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập ID %v", rc.ID), http.StatusUnprocessableEntity)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		rc.Purchase = rcp.Purchase

		err = ap.data.SetReceive(&rc)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Phiếu nhập ID %v có thể đã hoặc đang được điều chỉnh bởi 1 tài khoản khác. Hãy tải lại trang và thử lại", rc.ID), http.StatusUnprocessableEntity)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, "receive/"+rc.ID, http.StatusSeeOther)
	})
}

func (ap *application) delReceive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		rc, err := ap.data.Receive(id)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) || errors.Is(err, data.ErrInvalidID) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập ID %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		err = ap.data.DelReceive(rc.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rcs, err := ap.data.ReceivesByPurchase(rc.Purchase.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(rcs) == 0 {
			err = ap.data.UpdatePurchaseStatus(rc.Purchase.ID, rc.Purchase.Status, data.AwaitingResponse)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/receives", http.StatusSeeOther)
	})
}

func (ap *application) receivesPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v; %v", ErrConvertCtxVal, authenticatedCtxWarehouseID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rs, err := ap.data.Receives(wID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		tdata, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		tdata.Receives = rs

		err = ap.render(w, http.StatusOK, "receives", tdata)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) receivesByPurchasePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		purchaseID := r.PathValue("purchase")

		pc, err := ap.data.Purchase(purchaseID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rs, err := ap.data.ReceivesByPurchase(purchaseID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		tdata, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		tdata.Receives = rs
		tdata.Purchase = *pc

		err = ap.render(w, http.StatusOK, "receives_by_purchase", tdata)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) receiveProcessPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		rc, err := ap.data.Receive(id)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập ID: %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		if not01011000(rc.ActualAt) {
			ap.logger.Error(fmt.Sprintf("Receive %v already processed yet this req is made", rc.ID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Receive = *rc

		// td.ItemQuantitys, err = ap.data.UnreceivedPurchaseItemsOpt(rc)
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 	return
		// }

		err = ap.render(w, http.StatusOK, "receive_process", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) processReceive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rc data.Receive

		err := ap.decodeJSON(w, r, &rc)
		if err != nil {
			ap.logger.Error(err.Error())

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, http.StatusBadRequest)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Get receive to get the receive's purchase ID
		rcp, err := ap.data.Receive(rc.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// Transfer each receive item's notes
		for _, iq := range rc.Items {
			for i := range rcp.Items {
				if iq.GTIN == rcp.Items[i].GTIN {
					rcp.Items[i].Note = iq.Note
					break
				}
			}
		}

		// Update corresponding purchase status
		rs, err := ap.data.UnprocessedReceivesByPurchase(rcp.Purchase.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(rs) == 0 {
			err = ap.data.UpdatePurchaseStatus(rcp.Purchase.ID, rcp.Purchase.Status, data.Ended)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			err = ap.data.UpdatePurchaseStatus(rcp.Purchase.ID, rcp.Purchase.Status, data.Receiving)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// Set receive actual_at
		err = ap.data.SetActualAt(rcp)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Set receive processed_by
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v; %v", ErrConvertCtxVal.Error(), aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		rcp.ProcessedAccount.ID = aID
		err = ap.data.SetReceiveProcessedBy(rcp)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Set receive_item note
		err = ap.data.SetReceiveItemsNote(rcp)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Add serial to db
		for _, iq := range rc.Items {
			for _, s := range iq.Serials {
				s.Purchase.ID = rcp.Purchase.ID

				err = ap.data.AddSerial(&s)
				if err != nil {
					ap.logger.Error(err.Error())
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}
		}

		// http.Redirect(w, r, "/receives", http.StatusSeeOther)
		http.Redirect(w, r, fmt.Sprintf("%v/receive/%v/process/result", domain, rc.ID), http.StatusSeeOther)
	})
}

func (ap *application) receiveProcessResultPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		rc, err := ap.data.Receive(id)
		if err != nil {
			ap.logger.Error(err.Error(), ";", id)

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu nhập %v", id), http.StatusNotFound)
			} else {
				return
			}
		}

		err = ap.data.AddDifferenceSerialsByGTINOfPutawayReceive(rc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := ReceiveProcessResultPage{
			Receive: rc,
		}

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = ap.render(w, http.StatusOK, "receive_process_result", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) receive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		rc, err := ap.data.Receive(id)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập ID: %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		err = ap.writeJSON(w, http.StatusOK, rc, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

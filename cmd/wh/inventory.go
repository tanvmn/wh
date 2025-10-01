package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
)

func (ap *application) addInventoryPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		is, err := ap.data.NotExportedStockItems(wID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(is) == 0 {
			http.Error(w, fmt.Sprintf("Kho %v hiện không có hàng", wID), http.StatusNotFound)
			return
		}

		p := new(InventoryAddPage)
		p.Items = is

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = ap.render(w, http.StatusOK, "inventory_add", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) addInventory() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ExpectedAt string              `json:"expectedAt,omitempty,omitzero"`
			Items      []data.ItemQuantity `json:"items,omitempty,omitzero"`
		}

		err := ap.decodeJSON(w, r, &req)
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

		inventoryAddRequest := new(data.Inventory)
		inventoryAddRequest.ExpectedAt = req.ExpectedAt

		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v; authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		inventoryAddRequest.Warehouse.ID = wID

		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v; authenticatedCtxID: %v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		inventoryAddRequest.Account.ID = aID

		inventoryAddRequest.Items = req.Items

		id, err := ap.data.AddInventory(inventoryAddRequest)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/inventory/%v", id), http.StatusSeeOther)
	})
}

func (ap *application) inventoryPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		i, err := ap.data.Inventory(id)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoInventories) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiên kiểm %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		p := new(InventoryPage)
		p.Inventory = i

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "inventory", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) inventoryProcessPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inventoryID := r.PathValue("id")

		iss, err := ap.data.UncheckedInventorySerialsOf1RandomBin(inventoryID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// If there aren't any unchecked inventory serials (aka no unchecked bins) response 404
		if len(iss) == 0 {
			http.Error(w, "all bins checked", http.StatusNotFound)
			return
		}

		i, err := ap.data.Inventory(inventoryID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err = ap.data.UpdateInventoryStartedAt(inventoryID); err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(InventoryProcessPage)
		p.Inventory = i
		p.UncheckedInventorySerials = iss

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "inventory_process", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) processInventoryBinResult() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var inventoryBinResult data.Inventory

		err := ap.decodeJSON(w, r, &inventoryBinResult)
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

		println(inventoryBinResult.ID)
		for _, is := range inventoryBinResult.InventorySerials {
			println(is.Serial.NanoID)
			println(is.Result)
			println(is.Note)
			println()
		}

		http.Error(w, "success", http.StatusBadRequest)
	})
}

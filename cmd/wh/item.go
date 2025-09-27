package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
)

func (ap *application) name(i data.Item) string {
	return fmt.Sprintf("%v, %v, màu %v, %v, %v", i.Type, i.Brand, i.Color, i.Size, i.Characteristic)
}

func (ap *application) itemsPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		is, err := ap.data.Items()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		for i := range is {
			is[i].Stock, err = ap.data.Stock(is[i].GTIN, wID)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		if is != nil {
			td.Items = is
		}

		err = ap.render(w, http.StatusOK, "items", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) items() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		is, err := ap.data.Items()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if is == nil {
			ap.logger.Error(data.ErrNoItems.Error())
			http.Error(w, "Không tìm thấy mặt hàng nào", http.StatusNotFound)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, is, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) itemsBySupplier() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sID := r.URL.Query().Get("supplier")

		sp, err := ap.data.Supplier(sID)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoSuppliers) {
				http.Error(w, fmt.Sprintf("Không tìm thấy kho hoặc kho không tồn tại, ID: %v", sID), http.StatusBadRequest)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		is, err := ap.data.ItemsBySupplier(sp.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if is == nil {
			s := fmt.Sprintf("Không tìm thấy hàng theo nhà cung cấp %v", sp.ID)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusNotFound)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, is, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) unsafeStocksPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		s, err := ap.data.UnsafeStocks(wID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(UnsafeStockPage)
		p.UnsafeStocks = s

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = ap.render(w, http.StatusOK, "unsafe_stock", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

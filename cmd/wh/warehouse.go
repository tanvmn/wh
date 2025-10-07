package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
)

func (ap *application) unusedTotes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warehouseID := r.PathValue("warehouse")

		ts, err := ap.data.UnusedTotes(warehouseID)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrInvalidID) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if len(ts) == 0 {
			s := fmt.Sprintf("Không tìm thấy tote chưa sử dụng trong kho %v", warehouseID)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusNotFound)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, ts, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) binsPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warehouseID := r.URL.Query().Get("warehouse")

		wh, err := ap.data.Warehouse(warehouseID)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoWarehouses) {
				http.Error(w, fmt.Sprintf("Không tìm thấy kho %v", warehouseID), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		bs, err := ap.data.CurrentBinsEmptyPercentage(wh.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p := new(BinsPage)
		p.Warehouse = wh
		p.Bins = bs

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "bins", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

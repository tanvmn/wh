package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanvmn/wh/internal/data"
)

func (a *app) unusedTotes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warehouseID := r.PathValue("warehouse")

		ts, err := a.data.UnusedTotes(warehouseID)
		if err != nil {
			a.log.Error(err.Error())

			if errors.Is(err, data.ErrInvalidID) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if len(ts) == 0 {
			s := fmt.Sprintf("Không tìm thấy tote chưa sử dụng trong kho %v", warehouseID)
			a.log.Error(s)
			http.Error(w, s, http.StatusNotFound)
			return
		}

		err = a.writeJSON(w, http.StatusOK, ts, nil)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) binsPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warehouseID := r.URL.Query().Get("warehouse")

		wh, err := a.data.Warehouse(warehouseID)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoWarehouses) {
				http.Error(w, fmt.Sprintf("Không tìm thấy kho %v", warehouseID), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		bs, err := a.data.CurrentBinsEmptyPercentage(wh.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p := new(BinsPage)
		p.Warehouse = wh
		p.Bins = bs

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "bins", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

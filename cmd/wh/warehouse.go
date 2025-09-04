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

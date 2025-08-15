package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
)

func (ap *application) suppliers() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ss, err := ap.data.Suppliers()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if ss == nil {
			ap.logger.Error(data.ErrNoSuppliers.Error())
			http.Error(w, "Không tìm thấy nhà cung cấp nào", http.StatusNotFound)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, ss, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

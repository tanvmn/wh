package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
)

func (ap *application) items() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

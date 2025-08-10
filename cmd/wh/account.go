package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
)

// account handles GET /account?id= and response a JSON of internal/data.Account
func (ap *application) account() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		ac, err := ap.data.Account(id)
		if errors.Is(err, data.ErrNoAccounts) {
			s := fmt.Sprintf("Account %v not found", id)
			ap.logger.Error(s)
			http.Error(w, "Không tìm thấy tài khoản "+id, http.StatusNotFound)
			return
		} else if errors.Is(err, data.ErrInvalidID) {
			ap.logger.Error(err.Error())
			http.Error(w, fmt.Sprintf("Tài khoản ID '%v' không hợp lệ", id), http.StatusBadRequest)
			return
		} else if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, ac, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

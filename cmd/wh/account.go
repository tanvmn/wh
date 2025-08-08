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

		// i is the integer part of the id
		i, err := ap.id64(id, data.AccountIDCode)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, fmt.Sprintf("Tài khoản ID %v không hợp lệ", id), http.StatusBadRequest)
			return
		}

		ac, err := ap.data.Account(i)
		if errors.Is(err, data.ErrNoAccounts) {
			s := fmt.Sprintf("Account ACC-%v not found", i)
			ap.logger.Error(s)
			http.Error(w, "Không tìm thấy tài khoản "+id, http.StatusNotFound)
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

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
		id, err := ap.validateIDWithCode(r.URL.Query().Get("id"), data.IDCodes()["account"])
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ac, err := ap.data.Account(id)
		if errors.Is(err, data.ErrNoAccounts) {
			s := fmt.Sprintf("Không tìm thấy tài khoản, ACC-%v", id)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusNotFound)
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

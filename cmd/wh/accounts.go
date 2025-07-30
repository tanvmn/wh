package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tanNguyen2220022/wh/internal/util"
)

func (ap *application) account() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac, err := ap.data.Account.Get(1)
		if errors.Is(err, sql.ErrNoRows) {
			s := fmt.Sprintf("Account ACC-%v not found", 1)
			ap.logger.Info(s)
			http.Error(w, s, http.StatusNotFound)
			return
		} else if err != nil {
			ap.logger.Error(util.ErrLine)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ac.BDate, err = util.FormatRFC3339(ac.BDate, time.DateOnly, ap.logger)
		if err != nil {
			ap.logger.Error(util.ErrLine)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, ac, nil)
		if err != nil {
			ap.logger.Error(util.ErrLine)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

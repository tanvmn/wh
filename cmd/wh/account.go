package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tanNguyen2220022/wh/internal/util"
)

var (
	ErrAccountNotFound = errors.New("account not found")
)

func (ap *application) account() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sID := r.URL.Query().Get("id")
		id, err := ap.parseID(sID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, ErrInvalidID.Error()+", ACC-"+sID, http.StatusBadRequest)
			return
		}

		ac, err := ap.data.Account(id)
		if errors.Is(err, sql.ErrNoRows) {
			s := fmt.Sprintf("%v, ACC-%v", ErrAccountNotFound.Error(), id)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusNotFound)
			return
		} else if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ac.BDate, err = util.FormatRFC3339(ac.BDate, time.DateOnly)
		if err != nil {
			ap.logger.Error(err.Error())
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

package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
)

func (ap *application) putawayPromptPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.render(w, http.StatusOK, "putaway_prompt", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) putawayPageBySerial() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := r.URL.Query().Get("serial")

		rc, err := ap.data.ReceiveBySerial(sr)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập với serial %v", sr), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, "/putaway/"+rc.ID, http.StatusSeeOther)
	})
}

func (ap *application) putawayPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rID := r.PathValue("receive")

		rc, err := ap.data.Receive(rID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		pbs, err := ap.data.PutawayBins(rc.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := PutawayPage{
			PutawayBins: pbs,
			Receive:     rc,
		}
		td.Page = p

		err = ap.render(w, http.StatusOK, "putaway", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

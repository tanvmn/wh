package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
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

		rc, err := ap.data.UnputawayReceiveBySerial(sr)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập chưa cất với serial %v", sr), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		if notProcessed(rc.ActualAt) || not01011000(rc.PutawayAt) {
			ap.logger.Error(fmt.Sprintf("Receive %v is NOT processed or ALREADY PUTAWAY but somehow there is an serial %v from it", rc.ID, sr))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		if notProcessed(rc.ActualAt) || not01011000(rc.PutawayAt) {
			ap.logger.Error(fmt.Sprintf("Receive %v is NOT PROCESSED or ALREADY PUTAWAY but somehow reached this page is reached", rc.ID))
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

func (ap *application) putaway() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rc data.Receive

		err := ap.decodeJSON(w, r, &rc)
		if err != nil {
			ap.logger.Error(err.Error())

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v;%v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		rc.PutawayAccount.ID = aID

		err = ap.data.Putaway(&rc) // rc was used to catch JSON, NOT was queried from db
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.data.AddDifferenceSerials(&rc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.data.DelUnputawaySerials(&rc)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%v/putaway/%v/result", domain, rc.ID), http.StatusSeeOther)
	})
}

func (ap *application) putawayResultPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("receive")

		rc, err := ap.data.Receive(id)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p, err := ap.newPutawayResultPageByReceive(rc)
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
		td.Page = p

		err = ap.render(w, http.StatusOK, "putaway_result", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

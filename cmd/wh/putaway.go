package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/util"
)

func (a *app) putawayPromptPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = a.render(w, http.StatusOK, "putaway_prompt", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) putawayPageBySerial() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := r.URL.Query().Get("serial")

		rc, err := a.data.UnputawayReceiveBySerial(sr)
		if err != nil {
			a.log.Error(err.Error())

			if errors.Is(err, data.ErrNoReceives) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập chưa cất với serial %v", sr), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		if notProcessed(rc.ActualAt) || not01011000(rc.PutawayAt) {
			a.log.Error(fmt.Sprintf("Receive %v is NOT processed or ALREADY PUTAWAY but somehow there is an serial %v from it", rc.ID, sr))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/putaway/"+rc.ID, http.StatusSeeOther)
	})
}

func (a *app) putawayPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rID := r.PathValue("receive")

		rc, err := a.data.Receive(rID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if notProcessed(rc.ActualAt) || not01011000(rc.PutawayAt) {
			a.log.Error(fmt.Sprintf("Receive %v is NOT PROCESSED or ALREADY PUTAWAY but somehow reached this page is reached", rc.ID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		pbs, err := a.data.PutawayBins(rc.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := PutawayPage{
			PutawayBins: pbs,
			Receive:     rc,
		}
		td.Page = p

		err = a.render(w, http.StatusOK, "putaway", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) putaway() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var putawayResult data.Receive

		err := a.decodeJSON(w, r, &putawayResult)
		if err != nil {
			a.log.Error(err.Error())

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
			a.log.Error(fmt.Sprintf("%v;%v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		putawayResult.PutawayAccount.ID = aID

		err = a.data.Putaway(&putawayResult) // rc was used to catch JSON, NOT was queried from db
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = a.data.AddPutawayDifferenceSerials(&putawayResult)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = a.data.DelUnputawaySerials(&putawayResult)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%v/putaway/%v/result", domain, putawayResult.ID), http.StatusSeeOther)
	})
}

func (a *app) putawayResultPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("receive")

		rec, err := a.data.Receive(id)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// p, err := ap.newPutawayResultPageByReceive(rc)
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 	return
		// }

		for i := range rec.Items {
			it := &rec.Items[i]
			putaway, err := a.data.SuccessfullyPutawayQuantityByGTINAndReceive(rec.ID, it.Item.GTIN)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			it.SuccessfullyPutawayQuantity = putaway

			notPutaway, err := a.data.UnsuccessfullyPutawayQuantityByGTINAndReceive(rec.ID, it.Item.GTIN)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			it.UnsuccessfullyPutawayQuantity = notPutaway

			it.NeededPutawayQuantity = putaway + notPutaway
		}

		p := new(PutawayResultPage)
		p.Receive = rec

		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = a.render(w, http.StatusOK, "putaway_result", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

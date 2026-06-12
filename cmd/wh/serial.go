package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/util"
)

func (a *app) serialsPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID", ErrConvertCtxVal))
		}

		gtin := r.URL.Query().Get("gtin")

		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ss, err := a.data.SerialsByGTINAndWarehouse(gtin, wID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Serials = ss

		it, err := a.data.Item(gtin)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Item = *it

		err = a.render(w, http.StatusOK, "serials", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) outOfDateSerialsPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID", ErrConvertCtxVal))
		}

		gtin := r.URL.Query().Get("gtin")

		p := new(OutOfDateSerialsPage)

		is, err := a.data.OutOfDateItems(wID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		for _, iq := range is {
			if iq.Item.GTIN == gtin {
				for i := range iq.Serials {
					t, err := util.FormatRFC3339(iq.Serials[i].Receive.ActualAt, util.DDMMYYYY24HMI)
					if err != nil {
						a.log.Error(err.Error())
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					iq.Serials[i].Receive.ActualAt = t
				}

				p.Item = &iq
				break
			}
		}

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "outofdate_serials", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) serialsByBinPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		binID := r.URL.Query().Get("bin")

		b, err := a.data.Bin(binID)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoBins) {
				http.Error(w, fmt.Sprintf("Không tìm thấy Bin %v", binID), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		ss, err := a.data.SerialsByBin(binID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(SerialsByBinPage)
		p.Bin = b
		p.Serials = ss

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "serials_by_bin", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

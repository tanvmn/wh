package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
)

func (ap *application) addExport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resupplyID := r.URL.Query().Get("resupply")

		exportID, err := ap.data.AddExport(resupplyID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.data.SetResupplyStatus(resupplyID, data.AwaitingExport)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Error(w, exportID, http.StatusBadRequest)
	})
}

func (ap *application) exportPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		p := new(ExportPage)
		e, err := ap.data.Export(id)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		p.Export = e

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "export", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

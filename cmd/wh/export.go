package main

import (
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

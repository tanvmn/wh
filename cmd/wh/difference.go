package main

import (
	"fmt"
	"net/http"
)

func (a *app) differenceActivitiesPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v; %v", ErrConvertCtxVal, wID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		as, err := a.data.DifferenceActivities(wID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		dap := DifferenceActivitiesPage{
			DifferenceActivities: as,
		}

		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = dap

		err = a.render(w, http.StatusOK, "difference_activities", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

package main

import (
	"net/http"
)

func (ap *application) homePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error("Cannot convert authenticatedCtxID from any to string")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		role, ok := r.Context().Value(authenticatedCtxRole).(string)
		if !ok {
			ap.logger.Error("Cannot convert authenticatedCtxRole from any to string")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		td := ap.newTemplData(r)
		td.Account.ID = id
		td.Account.Role = role

		err := ap.render(w, http.StatusOK, "home", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

package main

import (
	"net/http"
)

func (ap *application) homePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ap.render(w, http.StatusOK, "home", templData{})
	})
}

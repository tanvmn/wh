package main

import (
	"net/http"
)

func (ap *application) homePage(rw http.ResponseWriter, rq *http.Request) {
	ap.render(rw, http.StatusOK, "home", templData{})
}

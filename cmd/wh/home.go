package main

import (
	"net/http"
)

func (a *application) homePage(w http.ResponseWriter, q *http.Request) {
	a.render(w, http.StatusOK, "home", templData{})
}

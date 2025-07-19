package main

import (
	"html/template"
	"net/http"
)

func (ap *application) homePage(rw http.ResponseWriter, rq *http.Request) {
	tp, err := template.ParseFiles("./ui/html/base.tmpl.html", "./ui/html/page/home.tmpl.html")
	if err != nil {
		ap.logger.Error(err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = tp.ExecuteTemplate(rw, "base", nil); err != nil {
		ap.logger.Error(err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func (a *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data templData) {
	tmpl, exist := a.templCache[page]
	if !exist {
		s := fmt.Sprintf("template %v does not exist", page)
		a.logger.Error(s)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	bf := new(bytes.Buffer)

	// write to a temp buffer first, if there's an error, response 500,
	err := tmpl.ExecuteTemplate(bf, "base", data)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// if there aren't any errors, write from temp buffer to ResponseWriter
	w.WriteHeader(status)
	bf.WriteTo(w)
}

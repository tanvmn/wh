package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func (ap *application) render(
	rw http.ResponseWriter,
	status int,
	page string,
	data templData) {
	tmpl, exist := ap.templCache[page]
	if !exist {
		s := fmt.Sprintf("template %v does not exist", page)
		ap.logger.Error(s)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	bf := new(bytes.Buffer)

	// write to a temp buffer first in case there's an error
	err := tmpl.ExecuteTemplate(bf, "base", data)
	if err != nil {
		ap.logger.Error(err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// if there aren't any errors, write from temp buffer to ResponseWriter
	rw.WriteHeader(status)
	bf.WriteTo(rw)
}

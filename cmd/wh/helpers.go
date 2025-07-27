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
	data templData) error {
	tmpl, exist := ap.templCache[page]
	if !exist {
		err := fmt.Errorf("template %v does not exist", page)
		ap.logger.Error(err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	bf := new(bytes.Buffer)

	// write to a temp buffer first in case there's an error
	err := tmpl.ExecuteTemplate(bf, "base", data)
	if err != nil {
		ap.logger.Error(err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	// if there aren't any errors, write from temp buffer to ResponseWriter
	rw.WriteHeader(status)
	bf.WriteTo(rw)

	return nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
)

func (ap *application) servePage(
	w http.ResponseWriter,
	status int,
	page string,
	data templData,
) error {
	tmpl, exist := ap.templCache[page]
	if !exist {
		err := fmt.Errorf("template '%v' does not exist", page)
		ap.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	buf := new(bytes.Buffer)

	// write to a temp buffer first in case there's an error
	err := tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
		ap.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	// if there aren't any errors, write from temp buffer to ResponseWriter
	w.WriteHeader(status)
	buf.WriteTo(w)

	return nil
}

func (ap *application) writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
	h http.Header,
) error {
	js, err := json.Marshal(data)
	if err != nil {
		ap.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	// add to or replace exsting k/v in response's headers
	maps.Copy(w.Header(), h)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

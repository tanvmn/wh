package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strconv"
)

var (
	// ErrIDLessThan1 = errors.New("ID less than 1")
	ErrInvalidID = errors.New("ID không hợp lệ")
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
		return err
	}

	// add to or replace exsting k/v in response's headers
	maps.Copy(w.Header(), h)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// parseID, if err == nil, return an id >= 1
func (ap *application) parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("ID ACC-%v không hợp lệ", s)
	}

	return id, nil
}
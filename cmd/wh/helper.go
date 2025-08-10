package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strconv"

	"github.com/tanNguyen2220022/wh/internal/validator"
)

const (
	companyName = "Morgan Maxwell"
	domain = "http://localhost:4000"
	itemImgPathFS = "item/img/"
)

func (ap *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data templData,
) error {
	tmpl, exist := ap.templCache[page]
	if !exist {
		err := fmt.Errorf("template '%v' does not exist", page)
		return err
	}

	buf := new(bytes.Buffer)

	// write to a temp buffer first in case there's an error
	err := tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
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

// id64 checks if id is at least 5 chars and if the code part is within permittedCodes,
// then parses the number part to an int64
func (ap *application) id64(id string, permittedCodes ...string) (int64, error) {
	va := validator.Validator{}

	va.Check(
		validator.MinChars(id, 5) && validator.Permitted(id[:4], permittedCodes...),
		fmt.Sprintf("ID %v is less than 5 chars or the code is not within %v", id, permittedCodes),
	)
	if !va.Valid() {
		return 0, errors.New(va.Message())
	}

	i, err := strconv.ParseInt(id[4:], 10, 64)
	if err != nil || i < 1 {
		return 0, fmt.Errorf("ID %v, the number must be >= 1", id)
	}

	return i, nil
}

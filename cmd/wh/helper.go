package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/validator"
)

var (
	// ErrIDLessThan1 = errors.New("ID less than 1")
	ErrInvalidID = errors.New("ID không hợp lệ")
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

// validateID validate the code part and parses the integer into an int64
func (ap *application) validateID(s string) (int64, error) {
	// codeValid := false
	// for _, v := range data.IDCodes() {
	// 	if s[:4] == v {
	// 		codeValid = true
	// 		break
	// 	}
	// }
	// if !codeValid {
	// 	return 0, fmt.Errorf("ID %v không hợp lệ", s)
	// }
	va := validator.Validator{}
	va.Check(
		validator.MinChars(s, 4) && validator.Permitted(s[:4], slices.Collect(maps.Values(data.IDCodes()))...),
		fmt.Sprintf("ID %v, mã %v không tồn tại", s, s[:4]),
	)
	if !va.Valid() {
		return 0, errors.New(va.Message())
	}

	id, err := strconv.ParseInt(s[4:], 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("ID %v, số %v không hợp lệ", s, s[4:])
	}

	return id, nil
}

func (ap *application) validateIDWithCode(s, code string) (int64, error) {
	va := validator.Validator{}
	va.Check(validator.MinChars(s, 4) && validator.Permitted(s[:4], code), fmt.Sprintf("ID %v, mã %v không phải Account ID", s, s[:4]))
	if !va.Valid() {
		return 0, errors.New(va.Message())
	}
	// if s[:4] != code {
	// 	return 0, fmt.Errorf("%v không thuộc mã %v", s, code)
	// }

	id, err := ap.validateID(s)
	if err != nil {
		return 0, err
	}

	return id, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/util"
)

const (
	companyName   = "Morgan Maxwell"
	domain        = "http://localhost:4000"
	itemImgPathFS = "item/img/"
)

var ErrConvertCtxVal = errors.New("cannnot convert context value to desired type")

func (a *app) render(
	w http.ResponseWriter,
	status int,
	page string,
	data templData,
) error {
	tmpl, exist := a.templCache[page]
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)

	return nil
}

func (a *app) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	maps.Copy(w.Header(), headers)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(js); err != nil {
		return err
	}

	return nil
}

func (a *app) json(w http.ResponseWriter, status int, data any, headers http.Header) {
	if err := a.writeJSON(w, status, data, headers); err != nil {
		a.log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *app) jsonOK(w http.ResponseWriter, data any) {
	a.json(w, http.StatusOK, data, nil)
}

// decodeJSON decode JSON body, slog errors, and returns client error code and message
func (a *app) decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// If the Content-Type header is present, get the value,
	// remove additional parameters like charset or boundary information,
	// and normalize by stripping whitespace and converting to lowercase before checking if it's application/json
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type is NOT application/json"
			a.log.Error(msg)
			return &util.MalformedRequest{
				Status: http.StatusUnsupportedMediaType,
				Msg:    msg,
			}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 2<<20) // (2 Bytes << 20) == 2 MBytes

	de := json.NewDecoder(r.Body)
	// de.DisallowUnknownFields()

	err := de.Decode(&dst)
	if err != nil {
		var (
			syntaxErr             *json.SyntaxError
			unmarshalTypeErr      *json.UnmarshalTypeError
			maxBytesErr           *http.MaxBytesError
			invalidUnmarshalError *json.InvalidUnmarshalError
		)

		switch {
		// Catch JSON syntax errors, and include the position of the error in JSON string
		case errors.As(err, &syntaxErr):
			a.log.Error(syntaxErr.Error(), "position", syntaxErr.Offset)
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    fmt.Sprintf("Malformed JSON at position %v", syntaxErr.Offset),
			}

		// Decode() may returns io.ErrUnexpectedError for JSON syntaxt error
		case errors.Is(err, io.ErrUnexpectedEOF):
			a.log.Error(err.Error())
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Malformed JSON",
			}

		// json.UnmarshalTypeError occurs when the JSON value is the wrongtype for the target destination
		case errors.As(err, &unmarshalTypeErr):
			a.log.Error(unmarshalTypeErr.Error(), "field", unmarshalTypeErr.Field, "position", unmarshalTypeErr.Offset, "type", unmarshalTypeErr.Type, "val", unmarshalTypeErr.Value)
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    fmt.Sprintf("JSON has invalid value for %q field at position %v", unmarshalTypeErr.Field, unmarshalTypeErr.Offset),
			}

		// Catch any unknown fields in JSON compared to the destination target
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			a.log.Error(err.Error(), "field", field)
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    fmt.Sprintf("JSON has unknown field %v", field),
			}

		// Occurs when JSON is empty
		case errors.Is(err, io.EOF):
			a.log.Error(err.Error())
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Request cannot be empty",
			}

		case errors.As(err, &maxBytesErr):
			a.log.Error(maxBytesErr.Error(), "bytes", maxBytesErr.Limit)
			return &util.MalformedRequest{
				Status: http.StatusRequestEntityTooLarge,
				Msg:    fmt.Sprintf("JSON cannot be larger than %v", maxBytesErr.Limit),
			}

		// Occurs when a nil pointer or something Go deems invalid, is passed to json.Unmarhshal
		case errors.As(err, &invalidUnmarshalError):
			a.log.Error(err.Error())
			return err

		default:
			a.log.Error(err.Error())
			return err
		}
	}

	err = de.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "JSON must be an array or has only 1 object"
		a.log.Error(msg)
		return &util.MalformedRequest{
			Status: http.StatusBadRequest,
			Msg:    msg,
		}
	}

	return nil
}

func (a *app) background(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				a.log.Error(fmt.Sprint(err))
				debug.PrintStack()
			}
		}()

		fn()
	}()
}

func (a *app) authenticatedAccount(r *http.Request) (*data.Account, error) {
	aID, ok := r.Context().Value(authenticatedCtxID).(string)
	if !ok {
		return nil, fmt.Errorf("%w, authenticatedCtxID %v", ErrConvertCtxVal, aID)
	}

	ac, err := a.data.Account(aID)
	if err != nil {
		return nil, err
	}

	return ac, nil
}

func (a *app) authenticatedWarehouse(r *http.Request) (*data.Warehouse, error) {
	wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
	if !ok {
		return nil, fmt.Errorf("%w, authenticatedCtxWarehouseID %v", ErrConvertCtxVal, wID)
	}

	wh, err := a.data.Warehouse(wID)
	if err != nil {
		return nil, err
	}

	return wh, nil
}

func (a *app) authenticatedStore(r *http.Request) (*data.Store, error) {
	sID, ok := r.Context().Value(authenticatedCtxStoreID).(string)
	if !ok {
		return nil, fmt.Errorf("%w, authenticatedCtxStoreID %v", ErrConvertCtxVal, sID)
	}

	st, err := a.data.Store(sID)
	if err != nil {
		return nil, err
	}

	return st, nil
}

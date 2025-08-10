package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/tanNguyen2220022/wh/internal/util"
)

const (
	companyName   = "Morgan Maxwell"
	domain        = "http://localhost:4000"
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

// id64 checks if id is at least 5 chars and if the code part is one of permittedCodes,
// then parses the number part to an int64
// func (ap *application) id64(id string, permittedCodes ...string) (int64, error) {
// 	va := validator.Validator{}

// 	va.Check(
// 		validator.MinChars(id, 5) && validator.Permitted(id[:4], permittedCodes...),
// 		fmt.Sprintf("ID %v is less than 5 chars or the code is not within %v", id, permittedCodes),
// 	)
// 	if !va.Valid() {
// 		return 0, errors.New(va.Message())
// 	}

// 	i, err := strconv.ParseInt(id[4:], 10, 64)
// 	if err != nil || i < 1 {
// 		return 0, fmt.Errorf("ID %v, the number must be >= 1", id)
// 	}

// 	return i, nil
// }

// decodeJSON decode JSON body, slog errors, and returns client error message
func (ap *application) decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type is NOT application/json"
			ap.logger.Error(msg)
			return &util.MalformedRequest{
				Status: http.StatusUnsupportedMediaType,
				Msg:    msg,
			}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 2<<20) // 2Bytes<<20 == 2 MBytes

	de := json.NewDecoder(r.Body)
	// de.DisallowUnknownFields()

	err := de.Decode(&dst)
	if err != nil {
		var (
			syntaxErr        *json.SyntaxError
			unmarshalTypeErr *json.UnmarshalTypeError
			maxBytesErr      *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxErr):
			ap.logger.Error(syntaxErr.Error(), "position", syntaxErr.Offset)
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    fmt.Sprintf("Malformed JSON at position %v", syntaxErr.Offset),
			}

		case errors.Is(err, io.ErrUnexpectedEOF):
			ap.logger.Error(err.Error())
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Malformed JSON",
			}

		case errors.As(err, &unmarshalTypeErr):
			ap.logger.Error(unmarshalTypeErr.Error(), "field", unmarshalTypeErr.Field, "position", unmarshalTypeErr.Offset, "type", unmarshalTypeErr.Type, "val", unmarshalTypeErr.Value)
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    fmt.Sprintf("JSON has invalid value for %q field at position %v", unmarshalTypeErr.Field, unmarshalTypeErr.Offset),
			}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			ap.logger.Error(err.Error(), "field", field)
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    fmt.Sprintf("JSON has unknown field %v", field),
			}

		case errors.Is(err, io.EOF):
			ap.logger.Error(err.Error())
			return &util.MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Request cannot be empty",
			}

		case errors.As(err, &maxBytesErr):
			ap.logger.Error(maxBytesErr.Error(), "bytes", maxBytesErr.Limit)
			return &util.MalformedRequest{
				Status: http.StatusRequestEntityTooLarge,
				Msg:    fmt.Sprintf("JSON cannot be larger than %v", maxBytesErr.Limit),
			}

		default:
			ap.logger.Error(err.Error())
			return err
		}
	}

	err = de.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "JSON must be an array or has 1 object"
		ap.logger.Error(msg)
		return &util.MalformedRequest{
			Status: http.StatusBadRequest,
			Msg:    msg,
		}
	}

	return nil
}

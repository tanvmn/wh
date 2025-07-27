package main

import (
	"encoding/json"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/util"
)

func (ap *application) homePage(rw http.ResponseWriter, rq *http.Request) {
	err := ap.render(rw, http.StatusOK, "home", templData{})
	if err != nil {
		ap.logger.Error(util.ErrLine)
		return
	}
}

func (ap *application) writeJSON(
	rw http.ResponseWriter,
	status int,
	data any,
	headers http.Header,
) error {
	js, err := json.Marshal(data)
	if err != nil {
		ap.logger.Error(err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	// ranging from a nil slice, map won't throw an error
	for k, v := range headers {
		rw.Header()[k] = v
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(js)

	return nil
}

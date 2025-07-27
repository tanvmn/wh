package main

import (
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

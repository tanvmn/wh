package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/util"
)

func (ap *application) homePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := ap.servePage(w, http.StatusOK, "home", templData{})
		if err != nil {
			ap.logger.Error(util.ErrLine)
			return
		}
	})
}

package main

import (
	"errors"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
)

func (ap *application) addPurchase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pc data.Purchase

		err := ap.decodeJSON(w, r, &pc)
		if err != nil {
			ap.logger.Error(util.ErrLine)

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		// fmt.Printf("%+v\n", pc)
	})
}

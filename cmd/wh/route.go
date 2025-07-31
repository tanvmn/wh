package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/rec"
	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/t", func(w http.ResponseWriter, r *http.Request) {
		// paras, err := url.ParseQuery(r.URL.RawQuery)
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	return
		// }

		paras := r.URL.Query()
		fmt.Println(paras)

		js, err := json.Marshal(paras)
		if err != nil {
			ap.logger.Error(err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(js))
	})

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", http.StripPrefix("/rec", http.FileServerFS(rec.Files)))

	mux.HandleFunc("GET /health", ap.health)
	mux.Handle("GET /{$}", ap.homePage())

	// Account
	mux.Handle("GET /account", ap.account())

	pre := middlewares{ap.recoverPanic, ap.logRequest, ap.addCommonHeaders}

	return pre.then(mux)
}

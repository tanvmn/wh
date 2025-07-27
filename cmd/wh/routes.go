package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/rec"
	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", http.StripPrefix("/rec", http.FileServerFS(rec.Files)))

	mux.HandleFunc("GET /health", ap.health)
	mux.HandleFunc("GET /{$}", ap.homePage)

	pre := middlewares{ap.recoverPanic, ap.logRequest, ap.addCommonHeaders}

	return pre.then(mux)
}

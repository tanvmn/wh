package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/ui"
)

func (a *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	// mux.Handle("GET /rec/", http.StripPrefix("/rec", http.FileServerFS(rec.Files)))

	mux.HandleFunc("GET /health", a.healthCheck)
	mux.HandleFunc("GET /{$}", a.homePage)

	preMux := middlewares{a.recoverPanic, a.logRequest, a.addCommondHeaders}

	return preMux.then(mux)
}

package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/rec"
	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", http.FileServerFS(rec.Files))

	mux.HandleFunc("GET /health", ap.healthCheck)
	mux.HandleFunc("GET /{$}", ap.homePage)

	return mux
}

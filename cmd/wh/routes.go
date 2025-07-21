package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/rec"
	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) routes() http.Handler {
	mx := http.NewServeMux()

	mx.Handle("GET /static/", http.FileServerFS(ui.Files))
	mx.Handle("GET /rec/", http.StripPrefix("/rec", http.FileServerFS(rec.Files)))

	mx.HandleFunc("GET /health", ap.healthCheck)
	mx.HandleFunc("GET /{$}", ap.homePage)

	return mx
}

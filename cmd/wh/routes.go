package main

import (
	"net/http"

	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) routes() http.Handler {
	sm := http.NewServeMux()

	sm.Handle("GET /static/", http.FileServerFS(ui.Files))

	sm.HandleFunc("GET /health", ap.healthCheck)
	sm.HandleFunc("GET /{$}", ap.homePage)

	return sm
}

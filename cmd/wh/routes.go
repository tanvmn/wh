package main

import (
	"net/http"
)

func (ap *application) routes() http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("GET /health", ap.healthCheck)
	sm.HandleFunc("GET /{$}", ap.homePage)

	return sm
}

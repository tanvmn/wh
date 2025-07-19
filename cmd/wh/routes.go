package main

import "net/http"

func (ap *application) routes() http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("GET /v1/health", ap.healthCheck)

	return sm
}

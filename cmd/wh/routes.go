package main

import "net/http"

func (ap *application) routes() http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("/v1/health", ap.healthCheck)

	return sm
}

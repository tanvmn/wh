package main

import "net/http"

func (a *app) err(w http.ResponseWriter, status int, data any) {
	a.json(w, status, map[string]any{"error": data}, nil)
}

func (a *app) tooManyRequests(w http.ResponseWriter) {
	a.err(w, http.StatusTooManyRequests, map[string]any{"error": "too many requests"})
}

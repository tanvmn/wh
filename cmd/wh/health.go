package main

import (
	"net/http"
)

func (a *app) health(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": a.config.env,
		"version":     version,
	}
	a.jsonOK(w, data)
}

package main

import (
	"fmt"
	"net/http"
)

func (ap *application) health(w http.ResponseWriter, r *http.Request) {
	js := `{"status":"available","environment":"%v","version":"%v"}`
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, js, ap.config.env, version)
}

package main

import (
	"fmt"
	"net/http"
)

func (ap *application) health(rw http.ResponseWriter, rq *http.Request) {
	js := `{"status":"available","environment":"%v","version":"%v"}`
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, js, ap.config.env, version)
}

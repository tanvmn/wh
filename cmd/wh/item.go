package main

import "net/http"

func (ap *application) items() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/items"))
	})
}

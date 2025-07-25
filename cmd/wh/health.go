package main

import (
	"fmt"
	"net/http"
)

func (a *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintln(w, "enviroment:", a.config.env)
	fmt.Fprintln(w, "version:", version)
}

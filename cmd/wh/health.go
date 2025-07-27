package main

import (
	"fmt"
	"net/http"
)

func (ap *application) health(rw http.ResponseWriter, rq *http.Request) {
	fmt.Fprintln(rw, "status: available")
	fmt.Fprintln(rw, "enviroment:", ap.config.env)
	fmt.Fprintln(rw, "version:", version)
}

package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
)

type middlewares []func(http.Handler) http.Handler

func (m middlewares) then(final http.Handler) http.Handler {
	slices.Reverse(m)
	for i := range m {
		final = m[i](final)
	}

	return final
}

func (ap *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ap.logger.Error(fmt.Sprint(err))
				fmt.Println(string(debug.Stack()))
				rw.Header().Set("Connection", "Close")
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(rw, rq)
	})
}

func (ap *application) addCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		// w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		// w.Header().Set("X-Content-Type-Options", "nosniff")
		// w.Header().Set("X-Frame-Options", "deny")
		// w.Header().Set("X-XSS-Protection", "0")

		// w.Header().Set("Server", "Go")

		next.ServeHTTP(rw, rq)
	})
}

func (ap *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		ap.logger.Info("request", "ip", rq.RemoteAddr, "method", rq.Method, "uri", rq.URL.RequestURI(), "proto", rq.Proto)

		next.ServeHTTP(rw, rq)
	})
}

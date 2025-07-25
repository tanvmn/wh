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

func (a *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.logger.Error(fmt.Sprint(err))
				fmt.Println(string(debug.Stack()))
				w.Header().Set("Connection", "Close")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (a *application) addCommondHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		// w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		// w.Header().Set("X-Content-Type-Options", "nosniff")
		// w.Header().Set("X-Frame-Options", "deny")
		// w.Header().Set("X-XSS-Protection", "0")

		// w.Header().Set("Server", "Go")

		next.ServeHTTP(w, r)
	})
}

func (a *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.logger.Info("request", "ip", r.RemoteAddr, "method", r.Method, "uri", r.URL.RequestURI(), "proto", r.Proto)

		next.ServeHTTP(w, r)
	})
}

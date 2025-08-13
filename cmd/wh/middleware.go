package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"

	"github.com/tanNguyen2220022/wh/internal/data"
)

type middlewares []func(http.Handler) http.Handler

func (ms middlewares) then(final http.Handler) http.Handler {
	for _, m := range slices.Backward(ms) {
		final = m(final)
	}

	return final
}

func (ap *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ap.logger.Error(fmt.Sprint(err))
				fmt.Println(string(debug.Stack()))
				w.Header().Set("Connection", "Close")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (ap *application) addHeaders(next http.Handler) http.Handler {
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

func (ap *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ap.logger.Info("request", "ip", r.RemoteAddr, "method", r.Method, "uri", r.URL.RequestURI(), "proto", r.Proto)

		next.ServeHTTP(w, r)
	})
}

func (ap *application) identify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := ap.sessionsManager.GetString(r.Context(), "authenticatedID")
		if id == "" {
			if r.URL.RequestURI() == "/login" {
				next.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}

		// Get account infos and store in request's context
		ac, err := ap.data.Account(id)
		if err != nil {
			if errors.Is(err, data.ErrNoAccounts) {
				ap.logger.Error(fmt.Sprintf("Account %v%v not found in db, but id is in session data", data.AccountIDCode, id))
				http.Error(w, "Tài khoản có thể không còn tồn tại từ sau phiên đăng nhập trước", http.StatusUnauthorized)
				return
			} else {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		if ac != nil {
			r = r.WithContext(context.WithValue(r.Context(), authenticatedCtxID, ac.ID))
			r = r.WithContext(context.WithValue(r.Context(), authenticatedCtxRole, ac.Role))
			r = r.WithContext(context.WithValue(r.Context(), authenticatedCtxWarehouseID, ac.Warehouse.ID))
			r = r.WithContext(context.WithValue(r.Context(), authenticatedCtxStoreID, ac.Store.ID))
		}

		// Set the "Cache-Control: no-store" header so that
		// pages require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")

		if r.URL.RequestURI() == "/login" && r.Method == http.MethodGet {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ap *application) permit(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(authenticatedCtxRole).(string)
			if !ok {
				ap.logger.Error("Cannot convert authenticatedCtxRole key to string")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if role != "Admin" && !slices.Contains(roles, role) {
				ap.logger.Warn(fmt.Sprintf(`Role "%v" cannot access this resource`, role))
				http.Error(w, fmt.Sprintf("Chức vụ %v không được truy cập vào tài nguyên này", role), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

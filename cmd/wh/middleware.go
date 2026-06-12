package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"sync"
	"time"

	"github.com/tanvmn/wh/internal/data"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

type middlewares []func(http.Handler) http.Handler

func (ms middlewares) then(final http.Handler) http.Handler {
	for _, m := range slices.Backward(ms) {
		final = m(final)
	}

	return final
}

func (a *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.log.Error(fmt.Sprint(err))
				fmt.Println(string(debug.Stack()))
				w.Header().Set("Connection", "Close")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (a *app) addHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		// w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		// w.Header().Set("X-Content-Type-Options", "nosniff")
		// w.Header().Set("X-Frame-Options", "deny")
		// w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (a *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.log.Info("request", "ip", r.RemoteAddr, "method", r.Method, "uri", r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (a *app) identify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := a.sessionsManager.GetString(r.Context(), "authenticatedID")
		if id == "" {
			if r.URL.RequestURI() == "/login" {
				next.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}

		// Get account infos and store in request's context
		ac, err := a.data.Account(id)
		if err != nil {
			if errors.Is(err, data.ErrNoAccounts) {
				a.log.Error(fmt.Sprintf("Account %v%v not found in db, but id is in session data", data.AccountIDCode, id))
				http.Error(w, "Tài khoản có thể không còn tồn tại từ sau phiên đăng nhập trước", http.StatusUnauthorized)
				return
			} else {
				a.log.Error(err.Error())
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

func (a *app) permit(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(authenticatedCtxRole).(string)
			if !ok {
				a.log.Error("Cannot convert authenticatedCtxRole key to string")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if role != "Admin" && !slices.Contains(roles, role) {
				a.log.Warn(fmt.Sprintf(`Role "%v" cannot access this resource`, role))
				http.Error(w, fmt.Sprintf("Chức vụ %v không được truy cập vào tài nguyên này", role), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (a *app) permitStoreEmployee(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(authenticatedCtxRole).(string)
		if !ok {
			a.log.Error("Cannot convert authenticatedCtxRole key to string")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if role != data.Admin {
			sI, ok := r.Context().Value(authenticatedCtxStoreID).(string)
			if !ok {
				a.log.Error(fmt.Sprintf("%v; authenticatedCtxStoreID %v", ErrConvertCtxVal, sI))
				http.Error(w, "Chỉ nhân viên cửa hàng truy cập được tài nguyên này", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (a *app) limitRate(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	newClient := func(ip string) *client {
		return &client{limiter: rate.NewLimiter(rate.Limit(a.config.limiter.rps), a.config.limiter.burst)}
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go a.background(func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.config.limiter.enabled {
			ip := realip.FromRequest(r)

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = newClient(ip)
			}
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				a.tooManyRequests(w)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

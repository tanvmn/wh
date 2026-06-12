package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *app) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", a.config.port),
		Handler:      a.routes(),
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(a.log.Handler(), slog.LevelError),
	}
	shutdownErr := make(chan error)

	// wait for SIGNINT or SIGTERM
	go a.background(func() {
		quit := make(chan os.Signal, 1)

		// will block until receiving a signal, but won't block sending to chan quit, so quit has to have at least 1 buffer.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		a.log.Info("shutting down server", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// return nil if Shutdown succeeds.
		// if returned error != nil, it's either from closing srv's Listener(s) or ctx expiring.
		shutdownErr <- srv.Shutdown(ctx)
	})

	a.log.Info(fmt.Sprintf("http://localhost:%v", a.config.port), "env", a.config.env)

	// calling Shutdown() on srv will cause ListenAndServe() to immediately returns http.ErrServerClosed, indicating graceful shutdown has started.
	// Therefore only return error if NOT http.ErrServerClosed.
	// ListenAndServe always returns a non-nil error.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// if err != nil, graceful shutdown failed.
	if err = <-shutdownErr; err != nil {
		return err
	}

	// at this point, graceful shutdown succeeded.
	a.log.Info("stopped server", "addr", srv.Addr)

	return nil
}

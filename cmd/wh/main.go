package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	dsn  string
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var cf config
	var err error

	flag.IntVar(&cf.port, "port", 4000, "HTTP server port")
	flag.StringVar(&cf.env, "env", "development", "Enviroment (development|staging|production)")
	flag.Parse()

	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))

	ap := &application{
		config: cf,
		logger: lg,
	}

	sv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cf.port),
		Handler:      ap.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(lg.Handler(), slog.LevelError),
	}

	lg.Info(fmt.Sprintf("http://localhost:%v", cf.port), "env", cf.env)
	err = sv.ListenAndServe()
	lg.Error(err.Error())
	os.Exit(1)
}

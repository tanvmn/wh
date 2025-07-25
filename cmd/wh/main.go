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
	dsn  string // data source name
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var c config
	var err error

	flag.IntVar(&c.port, "port", 4000, "HTTP server port")
	flag.StringVar(&c.env, "env", "development", "Enviroment (development|staging|production)")
	flag.Parse()

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))

	a := &application{
		config: c,
		logger: l,
	}

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.port),
		Handler:      a.routes(),
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(l.Handler(), slog.LevelError),
	}

	l.Info(fmt.Sprintf("http://localhost:%v", c.port), "env", c.env)
	err = s.ListenAndServe()
	l.Error(err.Error())
	os.Exit(1)
}

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tanNguyen2220022/wh/internal/util"
)

const version = "1.0.0"

type config struct {
	env  string
	dsn  string // data source name
	port int
}

type application struct {
	config     config
	logger     *slog.Logger
	templCache map[string]*template.Template
}

func main() {
	var cf config
	var err error

	flag.IntVar(&cf.port, "port", 4000, "HTTP server port")
	flag.StringVar(&cf.env, "env", "development", "Enviroment (development|staging|production)")
	flag.Parse()

	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))

	cache, err := newTemplCache(lg)
	if err != nil {
		lg.Error(util.ErrLine)
		os.Exit(1)
		return
	}
	a := &application{
		config:     cf,
		logger:     lg,
		templCache: cache,
	}

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", cf.port),
		Handler:      a.routes(),
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(lg.Handler(), slog.LevelError),
	}

	lg.Info(fmt.Sprintf("http://localhost:%v", cf.port), "env", cf.env)
	err = s.ListenAndServe()
	lg.Error(err.Error())
	os.Exit(1)
}

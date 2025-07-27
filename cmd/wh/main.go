package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/tanNguyen2220022/wh/internal/util"
)

const version = "1.0.0"

type config struct {
	env  string
	port int
	db   struct {
		dsn string // data source name
	}
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
	flag.StringVar(&cf.db.dsn, "dsn", "postgresql://wh:pa55word@localhost:5432/wh", "PostgreSQL DSN")
	flag.Parse()

	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))

	db, err := openDB(cf, lg)
	if err != nil {
		lg.Error(util.ErrLine)
		os.Exit(1)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			lg.Error(err.Error())
		}
	}()
	lg.Info("db connection pool established")

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

func openDB(cf config, lg *slog.Logger) (*sql.DB, error) {
	// create a empty connection pool
	db, err := sql.Open("postgres", cf.db.dsn)
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// try establishing a connection to db.
	// 1 case err != nil is a connection couldn't be established within 5 seconds
	err = db.PingContext(ctx)
	if err != nil {
		lg.Error(err.Error())
		db.Close()
		return nil, err
	}

	return db, nil
}

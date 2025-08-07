package main

import (
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/ui"
)

type templData struct {
	Domain  string
	Account struct {
		ID   string
		Role string
	}
	Items []data.Item
}

func (ap *application) newTemplData(r *http.Request) (templData, error) {
	id, ok := r.Context().Value(authenticatedCtxID).(string)
	if !ok {
		return templData{}, errors.New("error retrieving authenticated ID (string) from request's context")
	}
	role, ok := r.Context().Value(authenticatedCtxRole).(string)
	if !ok {
		return templData{}, errors.New("error retrieving authenticated ID (string) from request's context")
	}

	return templData{
		Domain: domain,
		Account: struct {
			ID   string
			Role string
		}{
			ID:   id,
			Role: role,
		},
	}, nil
}

func newTemplCache(lg *slog.Logger) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	// Get all the paths of the tmpl pages
	paths, err := fs.Glob(ui.Files, "html/pages/*.tmpl.html")
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}

	for _, path := range paths {
		// Get the *.tmpl.html part of the path, then the * part
		name := filepath.Base(path)
		name = name[:strings.Index(name, ".tmpl")]

		// Get the path's patterns of all tmpls needed for a page,
		// note that 'base' tmpl has to be the first element
		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.tmpl.html",
			path,
		}

		tmpl, err := template.ParseFS(ui.Files, patterns...)
		if err != nil {
			lg.Error(err.Error())
			return nil, err
		}

		cache[name] = tmpl
	}

	// Parse the login page
	cache["login"], err = template.ParseFS(ui.Files, "html/*.tmpl.html")
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}

	return cache, nil
}

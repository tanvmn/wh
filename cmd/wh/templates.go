package main

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/tanNguyen2220022/wh/ui"
)

type templData struct {
}

func (a *application) newTemplData(r *http.Request) templData {
	return templData{}
}

func newTemplCache(lg *slog.Logger) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	// get all the paths of the tmpl pages
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl.html")
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}

	for _, page := range pages {
		// get the *.tmpl.html part of the path, then get the * part of that
		name := filepath.Base(page)
		name = name[:strings.Index(name, ".tmpl")]

		// get the paths of all tmpls needed to make a page,
		// note that 'base' has to be the first element
		paths := []string{
			"html/base.tmpl.html",
			page,
		}

		tmpl, err := template.ParseFS(ui.Files, paths...)
		if err != nil {
			lg.Error(err.Error())
			return nil, err
		}

		cache[name] = tmpl
	}

	return cache, nil
}

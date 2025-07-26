package main

import (
	"html/template"
	"log/slog"
	"path/filepath"
	"strings"
)

type templData struct {
}

func newTemplCache(lg *slog.Logger) (map[string]*template.Template, error) {
	// cache := map[string]*template.Template{}
	cache := make(map[string]*template.Template)

	// get all the paths of the tmpl pages
	pages, err := filepath.Glob(filepath.Join(".", "ui", "html", "pages", "*.tmpl.html"))
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}

	for _, page := range pages {
		// get the *.tmpl.html part, then get the * of that
		name := filepath.Base(page)
		name = name[:strings.Index(name, ".tmpl")]

		// get path of all tmpls needed to make a page,
		// note that base has to be the first element
		files := []string{
			filepath.Join(".", "ui", "html", "base.tmpl.html"),
			page,
		}

		t, err := template.ParseFiles(files...)
		if err != nil {
			lg.Error(err.Error())
			return nil, err
		}

		cache[name] = t
	}

	return cache, nil
}

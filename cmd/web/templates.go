package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/pippimotta/snippet-box/internal/models"
	"github.com/pippimotta/snippet-box/ui"
)

type templateData struct {
	Snippet     *models.Snippet
	Snippets    []*models.Snippet
	CurrentYear int
	Form        any
	Flash       string
	IsAuthenticated bool
	CSRFToken string //Add a CSRFTokrn
}

func humanDate(t time.Time) string {
	if t.IsZero(){
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := fs.Glob(ui.Files,"html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)

		//Create a slice containing the filepath patterns for the templates
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}
		//use ParseFS instead of ParseFiles to parse all template files from the embedded filesystem
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

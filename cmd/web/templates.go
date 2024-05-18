package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/kvnloughead/contacts-app/internal/models"
	"github.com/kvnloughead/contacts-app/ui"
)

// Returns human readable date, formatted as 'DD Mon YYYY at hh:mm'.
// If t is the zero time, an empty string is returned.
// All non-empty times are converted to UTC.
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// template.FuncMap struct provides a string keyed map of template functions.
// Must be registered with the template before calling ParseFiles.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// Go templates only allow a single data argument, so we create a struct to
// store all necessary template data.
type templateData struct {
	CurrentYear     int
	Contact         models.Contact
	Contacts        []models.Contact
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	DeleteForm      bool
}

func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a map to serve as a cache.
	cache := map[string]*template.Template{}

	// Gather a slice of all `pages` templates from our embedded filesystem.
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	/* Grab the basename of each page, generate the corresponding template, and add it to the cache. */
	for _, page := range pages {
		name := filepath.Base(page)

		// Create a slice containing the necessary template filepaths patters.
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// Before parsing base into a template set we call New() and Funcs() to
		// register the template functions. ParseFS is used instead of ParseFiles,
		// for access to ui.Files.
		ts, err := template.New(name).Funcs(functions).ParseFS(
			ui.Files,
			patterns...,
		)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

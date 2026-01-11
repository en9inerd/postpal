package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/en9inerd/postpal/internal/config"
	"github.com/en9inerd/postpal/ui"
)

type templateData struct {
	Form        any
	CurrentYear int
	PageTitle   string
	PageDesc    string
	Config      *config.Config
}

type templateCache struct {
	templates map[string]*template.Template
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{}
}

func newTemplateCache() (*templateCache, error) {
	cache := &templateCache{
		templates: make(map[string]*template.Template),
	}

	tmplFS, err := fs.Sub(ui.Files, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to get templates subdirectory: %w", err)
	}

	pages, err := fs.Glob(tmplFS, "pages/*.tmpl.html")
	if err != nil {
		return nil, fmt.Errorf("failed to glob pages: %w", err)
	}

	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), ".tmpl.html")
		patterns := []string{
			"layouts/base.tmpl.html",
			"partials/*.tmpl.html",
			page,
		}

		ts, err := template.New("base").Funcs(templateFuncs()).ParseFS(tmplFS, patterns...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", page, err)
		}
		cache.templates[name] = ts
	}

	partials, err := fs.Glob(tmplFS, "partials/*.tmpl.html")
	if err != nil {
		return nil, fmt.Errorf("failed to glob partials: %w", err)
	}

	for _, partial := range partials {
		name := strings.TrimSuffix(filepath.Base(partial), ".tmpl.html")
		ts, err := template.New(name).Funcs(templateFuncs()).ParseFS(tmplFS, partial)
		if err != nil {
			return nil, fmt.Errorf("failed to parse partial %s: %w", partial, err)
		}
		cache.templates[name] = ts
	}

	return cache, nil
}

func (tc *templateCache) render(w http.ResponseWriter, name string, td *templateData) error {
	tmpl, ok := tc.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", td)
}

func (tc *templateCache) renderFragment(w http.ResponseWriter, name string, td *templateData) error {
	tmpl, ok := tc.templates[name]
	if !ok {
		return fmt.Errorf("template fragment %s not found", name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, name, td)
}

func renderError(w http.ResponseWriter, templates *templateCache, message string) {
	templates.renderFragment(w, "errors", &templateData{Form: map[string]string{"error": message}})
}

func renderPage(w http.ResponseWriter, logger *slog.Logger, templates *templateCache, pageName string, td *templateData) {
	if err := templates.render(w, pageName, td); err != nil {
		logger.Error("failed to render page", "page", pageName, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

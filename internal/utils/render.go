package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/justinas/nosurf"
	"html/template"
	"net/http"
	"path/filepath"
)

var functions = template.FuncMap{}

var app *config.AppConfig

var pathToTemplates = "./templates"

// AddDefaultData holds data that we want to send to our templates
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	// PopString puts a string into session until another page is displayed, and it removes itself then
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	// we add CSRF token to our default data
	td.CSRFToken = nosurf.Token(r)
	return td
}

// RenderTemplate checks the template cache we created earlier with CreateTemplateCache and renders the specific
// template that requested by user
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template

	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]

	if !ok {
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)

	if err != nil {
		fmt.Println("error writing template to browser:", err)
		return err
	}

	return nil
}

// CreateTemplateCache creates template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.gohtml", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))

		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	return myCache, nil
}

// NewTemplates takes memory address of the app variable that declared in main package and makes it reachable in this
// package
func NewTemplates(a *config.AppConfig) {
	app = a
}

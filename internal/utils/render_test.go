package utils

import (
	"github.com/burakkarasel/bookings/internal/models"
	"net/http"
	"testing"
)

// TestAddDefaultData test our func AddDefaultData in render.go
func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()

	if err != nil {
		t.Error(err)
	}

	// Here I put a key-value pair in my session
	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)

	// And if I cannot get that value from session my test fails
	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}

// TestRenderTemplate tests our RenderTemplate func in render.go
func TestTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()

	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()

	if err != nil {
		t.Error(err)
	}

	var ww myWriter
	// Here I used my myWriter type, so I can test my RenderTemplate func now

	err = Template(&ww, r, "home.page.gohtml", &models.TemplateData{})

	app.UseCache = true

	if err != nil {
		t.Error("error writing template to browser")
	}

	// To check if it fails when it's suppose to
	err = Template(&ww, r, "none-home.page.gohtml", &models.TemplateData{})

	if err == nil {
		t.Error("Rendered non-existing template")
	}
}

// TestNewTemplates Tests our NewTemplates func in render.go
func TestNewRenderer(t *testing.T) {
	NewRenderer(app)
}

// TestCreateTemplateCache test our CreateTemplateCache func in render.go
func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()

	if err != nil {
		t.Error(err)
	}
}

// getSession creates a http request with session, so we can test our AddDefaultData because we need session in it
func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)

	if err != nil {
		return nil, err
	}

	// here I reached our requests context and created a session in it and put it back in to my request
	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil
}

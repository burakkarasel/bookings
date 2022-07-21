package main

import (
	"testing"

	"github.com/burakkarasel/bookings/internal/config"
	"github.com/go-chi/chi"
)

// TestRoutes is test func for routes func in routes.go. It checks return type of func route.
func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		// do nothing
	default:
		t.Errorf("type is not *chi.Mux, type is %T", v)
	}
}

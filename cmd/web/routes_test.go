package main

import (
	"fmt"
	"github.com/JustAPotato0916/bookings/internal/config"
	"github.com/go-chi/chi/v5"
	"testing"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		// do nothing: test passed
	default:
		t.Error(fmt.Sprintf("type is not *chi.Muxm type is %T", v))
	}
}

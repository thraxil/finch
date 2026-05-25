package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
)

func TestHealthzHandler(t *testing.T) {
	// Setup dependencies
	p, cleanup := setupTestDB(t)
	defer cleanup()

	store := sessions.NewCookieStore([]byte("secret"))
	s := newSite(p, "http://localhost", store, "10", "true")

	// Create test server handler
	handler := NewServer("templates", "media", s, p)

	// Create request
	req, err := http.NewRequest("GET", "/healthz/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create ResponseRecorder
	rr := httptest.NewRecorder()

	// Serve HTTP
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

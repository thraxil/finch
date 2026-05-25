package main

import (
	"html/template"
	"testing"
	"time"
)

func TestPost(t *testing.T) {
	u := &user{
		ID:       1,
		Username: "testuser",
	}

	p := &post{
		ID:     1,
		UUID:   "1234-5678",
		User:   u,
		Body:   "Hello **world**",
		Posted: 1672531200, // 2023-01-01 00:00:00 UTC
	}

	// Test RenderBody
	rendered := p.RenderBody()
	expectedRendered := template.HTML("<p>Hello <strong>world</strong></p>\n")
	if rendered != expectedRendered {
		t.Errorf("RenderBody expected %q, got %q", expectedRendered, rendered)
	}

	// Test URL
	url := p.URL()
	expectedURL := "/u/testuser/p/1234-5678/"
	if url != expectedURL {
		t.Errorf("URL expected %q, got %q", expectedURL, url)
	}

	// Test Time
	expectedTime := time.Unix(1672531200, 0)
	if !p.Time().Equal(expectedTime) {
		t.Errorf("Time expected %v, got %v", expectedTime, p.Time())
	}
}

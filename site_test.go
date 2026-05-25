package main

import (
	"testing"

	"github.com/gorilla/sessions"
)

func TestSiteOperations(t *testing.T) {
	// Setup persistence
	p, cleanup := setupTestDB(t)
	defer cleanup()

	// Setup site
	store := sessions.NewCookieStore([]byte("something-very-secret"))
	s := newSite(p, "http://localhost:8000", store, "10", "true")

	// Create user via site
	u, err := s.CreateUser("siteuser", "sitepass")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if u.Username != "siteuser" {
		t.Errorf("Expected username siteuser, got %v", u.Username)
	}

	// Get user via site
	fetchedUser, err := s.GetUser("siteuser")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if fetchedUser.Username != "siteuser" {
		t.Errorf("Expected username siteuser, got %v", fetchedUser.Username)
	}

	// Add channels
	names := []string{"Site Channel", "Another Channel"}
	channels, err := s.AddChannels(*u, names)
	if err != nil {
		t.Fatalf("AddChannels failed: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("Expected 2 channels, got %d", len(channels))
	}

	// Get channel by ID and Slug
	fetchedChannel, err := s.GetChannelByID(channels[0].ID)
	if err != nil {
		t.Fatalf("GetChannelByID failed: %v", err)
	}
	if fetchedChannel.Label != "Site Channel" {
		t.Errorf("Expected label 'Site Channel', got %q", fetchedChannel.Label)
	}

	fetchedChannelBySlug, err := s.GetChannel(*u, "site_channel")
	if err != nil {
		t.Fatalf("GetChannel failed: %v", err)
	}
	if fetchedChannelBySlug.ID != channels[0].ID {
		t.Errorf("Expected ID %d, got %d", channels[0].ID, fetchedChannelBySlug.ID)
	}

	// Add Post
	post, err := s.AddPost(*u, "Site post body", channels)
	if err != nil {
		t.Fatalf("AddPost failed: %v", err)
	}

	// Verify post was returned
	if post.Body != "Site post body" {
		t.Errorf("Expected body 'Site post body', got %q", post.Body)
	}

	// Get all posts
	posts, err := s.GetAllPosts(10, 0)
	if err != nil {
		t.Fatalf("GetAllPosts failed: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("Expected 1 post, got %d", len(posts))
	}

	// Search posts
	searchPosts, err := s.SearchPosts("Site post", 10, 0)
	if err != nil {
		t.Fatalf("SearchPosts failed: %v", err)
	}
	if len(searchPosts) != 1 {
		t.Fatalf("Expected 1 search result, got %d", len(searchPosts))
	}

	// Delete Channel
	err = s.DeleteChannel(channels[1])
	if err != nil {
		t.Fatalf("DeleteChannel failed: %v", err)
	}

	userChannels, err := s.GetUserChannels(*u)
	if err != nil {
		t.Fatalf("GetUserChannels failed: %v", err)
	}
	if len(userChannels) != 1 {
		t.Fatalf("Expected 1 user channel after delete, got %d", len(userChannels))
	}
}

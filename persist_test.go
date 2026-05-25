package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*persistence, func()) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Read and execute schema.sql
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("Failed to read schema.sql: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to execute schema.sql: %v", err)
	}

	p := &persistence{Database: db}

	cleanup := func() {
		p.Close()
	}

	return p, cleanup
}

func TestPersistenceUser(t *testing.T) {
	p, cleanup := setupTestDB(t)
	defer cleanup()

	// Test CreateUser
	u, err := p.CreateUser("testuser", "password123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if u.Username != "testuser" {
		t.Errorf("Expected username testuser, got %v", u.Username)
	}

	// Test GetUser
	fetchedUser, err := p.GetUser("testuser")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if fetchedUser.Username != "testuser" {
		t.Errorf("Expected username testuser, got %v", fetchedUser.Username)
	}

	// Test check password
	if !fetchedUser.CheckPassword("password123") {
		t.Error("Password check failed for fetched user")
	}
}

func TestPersistenceChannelsAndPosts(t *testing.T) {
	p, cleanup := setupTestDB(t)
	defer cleanup()

	// Create user
	u, err := p.CreateUser("testuser", "password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Create channels
	names := []string{"General", "Random Thoughts"}
	channels, err := p.AddChannels(*u, names)
	if err != nil {
		t.Fatalf("AddChannels failed: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("Expected 2 channels, got %d", len(channels))
	}

	// Get user channels
	userChannels, err := p.GetUserChannels(*u)
	if err != nil {
		t.Fatalf("GetUserChannels failed: %v", err)
	}
	if len(userChannels) != 2 {
		t.Fatalf("Expected 2 user channels, got %d", len(userChannels))
	}

	// Create post with channels
	body := "Hello world in two channels"
	post, err := p.AddPost(*u, body, channels)
	if err != nil {
		t.Fatalf("AddPost failed: %v", err)
	}

	if post.Body != body {
		t.Errorf("Expected body %q, got %q", body, post.Body)
	}

	// Get post by UUID
	fetchedPost, err := p.GetPostByUUID(post.UUID)
	if err != nil {
		t.Fatalf("GetPostByUUID failed: %v", err)
	}
	if fetchedPost.ID != post.ID {
		t.Errorf("Expected post ID %d, got %d", post.ID, fetchedPost.ID)
	}

	// Get all posts
	posts, err := p.GetAllPosts(10, 0)
	if err != nil {
		t.Fatalf("GetAllPosts failed: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("Expected 1 post, got %d", len(posts))
	}
	if len(posts[0].Channels) != 2 {
		t.Errorf("Expected post to have 2 channels, got %d", len(posts[0].Channels))
	}

	// Test DeletePost
	err = p.DeletePost(post)
	if err != nil {
		t.Fatalf("DeletePost failed: %v", err)
	}

	postsAfterDelete, _ := p.GetAllPosts(10, 0)
	if len(postsAfterDelete) != 0 {
		t.Errorf("Expected 0 posts after deletion, got %d", len(postsAfterDelete))
	}
}

package handlers

import (
	"forum/database"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestCreatePostHandlerCreatesPost verifies post creation through the HTTP handler.
func TestCreatePostHandlerCreatesPost(t *testing.T) {
	// Start with a clean database and an author account.
	db := openTestDB(t)

	if err := database.CreateUser(db, "author@example.com", "author", "password"); err != nil {
		t.Fatal(err)
	}
	// Fetching the user confirms setup succeeded and gives the expected session ID.
	user, err := database.GetUserByEmail(db, "author@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}

	// Build a form submission matching the create-post endpoint.
	form := url.Values{
		"title":        {"First post"},
		"content":      {"Hello forum"},
		"category_ids": {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: "1"})
	w := httptest.NewRecorder()

	CreatePostHandler(db)(w, req)

	// Successful creation redirects to the feed.
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	// Read the database directly to confirm the handler inserted the post.
	posts, err := database.GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}
	// The clean database should contain only the submitted post.
	if len(posts) != 1 {
		t.Fatalf("posts = %d, want 1", len(posts))
	}
	// Assert the title passed through from form data to storage.
	if posts[0].Title != "First post" {
		t.Fatalf("title = %q, want First post", posts[0].Title)
	}
}

// TestCreatePostHandlerRejectsInvalidSession covers non-numeric session values.
func TestCreatePostHandlerRejectsInvalidSession(t *testing.T) {
	// The malformed cookie should be rejected before the database is touched.
	req := httptest.NewRequest(http.MethodPost, "/create-post", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bad-session"})
	w := httptest.NewRecorder()

	CreatePostHandler(nil)(w, req)

	// Invalid sessions are authentication failures.
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	// The response should identify the session parsing problem.
	if !strings.Contains(w.Body.String(), "invalid session") {
		t.Fatalf("body = %q, want invalid session message", w.Body.String())
	}
}

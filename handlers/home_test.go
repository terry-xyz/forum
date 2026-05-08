package handlers

import (
	"forum/database"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHomeHandlerRejectsUnsupportedMethod verifies HomeHandler is GET-only.
func TestHomeHandlerRejectsUnsupportedMethod(t *testing.T) {
	// A POST to home should be rejected before the database is used.
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(nil)(w, req)

	// Unsupported methods should return the standard 405 response.
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	// The response body should make the method problem obvious.
	if !strings.Contains(w.Body.String(), "method not allowed") {
		t.Fatalf("body = %q, want method-not-allowed message", w.Body.String())
	}
}

// TestHomeHandlerRejectsMissingSession verifies authentication is required.
func TestHomeHandlerRejectsMissingSession(t *testing.T) {
	// Do not attach a session cookie so the handler takes the unauthorized path.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(nil)(w, req)

	// Missing sessions should be rejected before any database query.
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// The body should contain the unauthorized message used by the handler.
	if !strings.Contains(w.Body.String(), "unauthorized") {
		t.Fatalf("body = %q, want unauthorized message", w.Body.String())
	}
}

// TestHomeHandlerRendersPostsWithAuthors verifies the happy-path home output.
func TestHomeHandlerRendersPostsWithAuthors(t *testing.T) {
	// Build a database with one author and one post.
	db := openTestDB(t)

	if err := database.CreateUser(db, "author@example.com", "author", "password"); err != nil {
		t.Fatal(err)
	}
	// Fetch the generated author ID so the post can reference it.
	user, err := database.GetUserByEmail(db, "author@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}
	if _, err := database.CreatePost(db, user.ID, "First post", "Hello forum"); err != nil {
		t.Fatal(err)
	}

	// Add a valid session cookie because HomeHandler requires authentication.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "1"})
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	// A valid request should render successfully.
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// The response should include both the post title and resolved author name.
	body := w.Body.String()
	if !strings.Contains(body, "<h3>First post</h3>") {
		t.Fatalf("body = %q, want rendered post title", body)
	}
	if !strings.Contains(body, "Author: author") {
		t.Fatalf("body = %q, want rendered author", body)
	}
}

// TestHomeHandlerReportsMissingCommentAuthor covers broken comment references.
func TestHomeHandlerReportsMissingCommentAuthor(t *testing.T) {
	// Build a valid post first so the only inconsistency is the comment author.
	db := openTestDB(t)

	if err := database.CreateUser(db, "author@example.com", "author", "password"); err != nil {
		t.Fatal(err)
	}
	user, err := database.GetUserByEmail(db, "author@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}
	postID, err := database.CreatePost(db, user.ID, "First post", "Hello forum")
	if err != nil {
		t.Fatal(err)
	}
	// Insert a comment with an author_id that does not exist to simulate bad data.
	if _, err := db.Exec("INSERT INTO comments (author_id, post_id, content) VALUES (?, ?, ?)", 999, postID, "orphaned comment"); err != nil {
		t.Fatal(err)
	}
	comments, err := database.GetCommentsByPostID(db, postID)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 {
		t.Fatalf("comments = %d, want 1", len(comments))
	}
	if comments[0].AuthorID != 999 {
		t.Fatalf("comment author id = %d, want 999", comments[0].AuthorID)
	}
	missingAuthor, err := database.GetUserByID(db, 999)
	if err != nil {
		t.Fatal(err)
	}
	if missingAuthor != nil {
		t.Fatalf("missing author = %#v, want nil", missingAuthor)
	}
	posts, err := database.GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts = %d, want 1", len(posts))
	}
	if posts[0].ID != postID {
		t.Fatalf("post id = %d, want %d", posts[0].ID, postID)
	}

	// Authenticate as the valid user so rendering reaches the comments block.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "1"})
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	// The handler has already written the filter bar before this late render
	// error, so httptest keeps the status at 200 while the body contains the
	// generic client-facing message.
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "failed to render posts") {
		t.Fatalf("body = %q, want render failure message", w.Body.String())
	}
}

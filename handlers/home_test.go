package handlers

import (
	"forum/database"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandlerRejectsUnsupportedMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(nil)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	if !strings.Contains(w.Body.String(), "method not allowed") {
		t.Fatalf("body = %q, want method-not-allowed message", w.Body.String())
	}
}

func TestHomeHandlerRejectsMissingSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(nil)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	if !strings.Contains(w.Body.String(), "unauthorized") {
		t.Fatalf("body = %q, want unauthorized message", w.Body.String())
	}
}

func TestHomeHandlerRendersPostsWithAuthors(t *testing.T) {
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
	if err := database.CreatePost(db, user.ID, "First post", "Hello forum"); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "1"})
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<h3>First post</h3>") {
		t.Fatalf("body = %q, want rendered post title", body)
	}
	if !strings.Contains(body, "Author: author") {
		t.Fatalf("body = %q, want rendered author", body)
	}
}

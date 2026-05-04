package handlers

import (
	"forum/database"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreatePostHandlerCreatesPost(t *testing.T) {
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

	form := url.Values{
		"title":   {"First post"},
		"content": {"Hello forum"},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: "1"})
	w := httptest.NewRecorder()

	CreatePostHandler(db)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	posts, err := database.GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts = %d, want 1", len(posts))
	}
	if posts[0].Title != "First post" {
		t.Fatalf("title = %q, want First post", posts[0].Title)
	}
}

func TestCreatePostHandlerRejectsInvalidSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/create-post", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bad-session"})
	w := httptest.NewRecorder()

	CreatePostHandler(nil)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "invalid session") {
		t.Fatalf("body = %q, want invalid session message", w.Body.String())
	}
}

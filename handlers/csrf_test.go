package handlers

import (
	"forum/database"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMutatingHandlersRejectMissingCSRFToken verifies protected writes require more than a session cookie.
func TestMutatingHandlersRejectMissingCSRFToken(t *testing.T) {
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
	sessionID := createSessionForUserID(t, db, "csrf-session", user.ID)

	tests := []struct {
		name    string
		path    string
		handler http.HandlerFunc
	}{
		{name: "comment", path: "/comment", handler: CreateCommentHandler(db)},
		{name: "post reaction", path: "/post-reaction", handler: ReactPostHandler(db)},
		{name: "comment reaction", path: "/comment-reaction", handler: ReactCommentHandler(db)},
		{name: "delete post", path: "/delete-post", handler: DeletePostHandler(db)},
		{name: "delete comment", path: "/delete-comment", handler: DeleteCommentHandler(db)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
			}
		})
	}
}

// TestHomeHandlerRendersCSRFToken verifies authenticated forms receive the session token.
func TestHomeHandlerRendersCSRFToken(t *testing.T) {
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
	if _, err := database.CreatePost(db, user.ID, "First post", "Hello forum"); err != nil {
		t.Fatal(err)
	}
	sessionID := createSessionForUserID(t, db, "home-csrf-session", user.ID)
	csrfToken := csrfTokenForTest(t, db, sessionID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !containsAll(w.Body.String(), `name="csrf_token"`, `value="`+csrfToken+`"`) {
		t.Fatalf("body = %q, want rendered csrf token", w.Body.String())
	}
}

func containsAll(s string, values ...string) bool {
	for _, value := range values {
		if !strings.Contains(s, value) {
			return false
		}
	}
	return true
}

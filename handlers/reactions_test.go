package handlers

import (
	"forum/database"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestReactPostHandlerRejectsMissingPost(t *testing.T) {
	db := openTestDB(t)
	_, sessionID, csrfToken := createAuthenticatedUserForTest(t, db, "reactor@example.com", "missing-post-reaction-session")

	form := url.Values{
		"post_id":       {"9999"},
		"reaction_type": {"like"},
		"csrf_token":    {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/post-reaction", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	ReactPostHandler(db)(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestReactCommentHandlerRejectsMissingComment(t *testing.T) {
	db := openTestDB(t)
	authorID, _, _ := createAuthenticatedUserForTest(t, db, "author@example.com", "comment-author-session")
	_, sessionID, csrfToken := createAuthenticatedUserForTest(t, db, "reactor@example.com", "missing-comment-reaction-session")
	postID, err := database.CreatePost(db, authorID, "First post", "Hello forum")
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"comment_id":    {strconv.Itoa(postID + 9999)},
		"reaction_type": {"like"},
		"csrf_token":    {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/comment-reaction", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	ReactCommentHandler(db)(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

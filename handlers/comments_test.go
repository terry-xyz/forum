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

func TestCreateCommentHandlerRejectsOversizedContent(t *testing.T) {
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
	sessionID := createSessionForUserID(t, db, "create-comment-session", user.ID)
	csrfToken := csrfTokenForTest(t, db, sessionID)

	form := url.Values{
		"post_id":    {strconv.Itoa(postID)},
		"content":    {strings.Repeat("C", maxCommentContentChars+1)},
		"csrf_token": {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	CreateCommentHandler(db)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "comment cannot exceed 280 characters") {
		t.Fatalf("body = %q, want comment length error", w.Body.String())
	}
	comments, err := database.GetCommentsByPostID(db, postID)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 0 {
		t.Fatalf("comments = %d, want 0", len(comments))
	}
}

func TestCreateCommentHandlerRejectsOversizedRequestBody(t *testing.T) {
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
	sessionID := createSessionForUserID(t, db, "create-comment-session", user.ID)
	csrfToken := csrfTokenForTest(t, db, sessionID)

	form := url.Values{
		"post_id":    {strconv.Itoa(postID)},
		"content":    {strings.Repeat("A", int(maxFormBodyBytes))},
		"csrf_token": {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	CreateCommentHandler(db)(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}
	comments, err := database.GetCommentsByPostID(db, postID)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 0 {
		t.Fatalf("comments = %d, want 0", len(comments))
	}
}

func TestDeleteCommentHandlerRejectsNonOwner(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "owner@example.com", "owner", "password"); err != nil {
		t.Fatal(err)
	}
	owner, err := database.GetUserByEmail(db, "owner@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if owner == nil {
		t.Fatal("owner was not created")
	}
	if err := database.CreateUser(db, "other@example.com", "other", "password"); err != nil {
		t.Fatal(err)
	}
	other, err := database.GetUserByEmail(db, "other@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if other == nil {
		t.Fatal("other user was not created")
	}

	postID, err := database.CreatePost(db, owner.ID, "First post", "Hello forum")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.CreateComment(db, owner.ID, postID, "owned comment"); err != nil {
		t.Fatal(err)
	}
	comments, err := database.GetCommentsByPostID(db, postID)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 {
		t.Fatalf("comments = %d, want 1", len(comments))
	}

	sessionID := createSessionForUserID(t, db, "delete-comment-session", other.ID)
	csrfToken := csrfTokenForTest(t, db, sessionID)
	form := url.Values{
		"comment_id": {strconv.Itoa(comments[0].ID)},
		"csrf_token": {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/delete-comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	DeleteCommentHandler(db)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	comments, err = database.GetCommentsByPostID(db, postID)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 {
		t.Fatalf("comments = %d, want 1", len(comments))
	}
	if comments[0].Content != "owned comment" {
		t.Fatalf("comment content = %q, want owned comment", comments[0].Content)
	}
}

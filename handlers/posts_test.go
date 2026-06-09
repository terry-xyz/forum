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
	sessionID := createSessionForUserID(t, db, "create-post-session", user.ID)
	csrfToken := csrfTokenForTest(t, db, sessionID)

	// Build a form submission matching the create-post endpoint.
	form := url.Values{
		"title":        {"First post"},
		"content":      {"Hello forum"},
		"category_ids": {"1"},
		"csrf_token":   {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
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

// TestCreatePostHandlerEscapesCategoryNames verifies category labels cannot inject HTML.
func TestCreatePostHandlerEscapesCategoryNames(t *testing.T) {
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
	if _, err := db.Exec("INSERT INTO categories (name) VALUES (?)", `<img src=x onerror=alert("category-xss")>`); err != nil {
		t.Fatal(err)
	}
	sessionID := createSessionForUserID(t, db, "create-post-session", user.ID)

	req := httptest.NewRequest(http.MethodGet, "/create-post", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	CreatePostHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	rawCategory := `<img src=x onerror=alert("category-xss")>`
	if strings.Contains(body, rawCategory) {
		t.Fatalf("body contains unescaped category name %q: %q", rawCategory, body)
	}
	escapedCategory := `&lt;img src=x onerror=alert(&#34;category-xss&#34;)&gt;`
	if !strings.Contains(body, escapedCategory) {
		t.Fatalf("body = %q, want escaped category name %q", body, escapedCategory)
	}
}

// TestCreatePostHandlerRejectsWhitespacePostFields verifies blank-looking posts cannot be created.
func TestCreatePostHandlerRejectsWhitespacePostFields(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		message string
	}{
		{
			name:    "blank title",
			title:   "   ",
			content: "Hello forum",
			message: "title cannot be empty",
		},
		{
			name:    "blank content",
			title:   "First post",
			content: "\t\n ",
			message: "content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			sessionID := createSessionForUserID(t, db, "create-post-session", user.ID)
			csrfToken := csrfTokenForTest(t, db, sessionID)

			form := url.Values{
				"title":        {tt.title},
				"content":      {tt.content},
				"category_ids": {"1"},
				"csrf_token":   {csrfToken},
			}
			req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
			w := httptest.NewRecorder()

			CreatePostHandler(db)(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			if !strings.Contains(w.Body.String(), tt.message) {
				t.Fatalf("body = %q, want %q", w.Body.String(), tt.message)
			}

			posts, err := database.GetAllPosts(db)
			if err != nil {
				t.Fatal(err)
			}
			if len(posts) != 0 {
				t.Fatalf("posts = %d, want 0", len(posts))
			}
		})
	}
}

func TestCreatePostHandlerRejectsOversizedFields(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		message string
	}{
		{
			name:    "oversized title",
			title:   strings.Repeat("T", maxPostTitleChars+1),
			content: "Hello forum",
			message: "title cannot exceed 280 characters",
		},
		{
			name:    "oversized content",
			title:   "First post",
			content: strings.Repeat("C", maxPostContentChars+1),
			message: "content cannot exceed 280 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			sessionID := createSessionForUserID(t, db, "create-post-session", user.ID)
			csrfToken := csrfTokenForTest(t, db, sessionID)

			form := url.Values{
				"title":        {tt.title},
				"content":      {tt.content},
				"category_ids": {"1"},
				"csrf_token":   {csrfToken},
			}
			req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
			w := httptest.NewRecorder()

			CreatePostHandler(db)(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			if !strings.Contains(w.Body.String(), tt.message) {
				t.Fatalf("body = %q, want %q", w.Body.String(), tt.message)
			}

			posts, err := database.GetAllPosts(db)
			if err != nil {
				t.Fatal(err)
			}
			if len(posts) != 0 {
				t.Fatalf("posts = %d, want 0", len(posts))
			}
		})
	}
}

func TestCreatePostHandlerRejectsOversizedRequestBody(t *testing.T) {
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
	sessionID := createSessionForUserID(t, db, "create-post-session", user.ID)
	csrfToken := csrfTokenForTest(t, db, sessionID)

	form := url.Values{
		"title":        {"First post"},
		"content":      {strings.Repeat("A", int(maxFormBodyBytes))},
		"category_ids": {"1"},
		"csrf_token":   {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	CreatePostHandler(db)(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}
	posts, err := database.GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatalf("posts = %d, want 0", len(posts))
	}
}

// TestCreatePostHandlerRejectsMissingCSRFToken verifies a valid session alone is not enough for writes.
func TestCreatePostHandlerRejectsMissingCSRFToken(t *testing.T) {
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
	sessionID := createSessionForUserID(t, db, "create-post-session", user.ID)

	form := url.Values{
		"title":        {"CSRF post"},
		"content":      {"This should not be created"},
		"category_ids": {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	CreatePostHandler(db)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	posts, err := database.GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatalf("posts = %d, want 0", len(posts))
	}
}

// TestCreatePostHandlerRejectsInvalidCSRFToken verifies forged tokens cannot authorize writes.
func TestCreatePostHandlerRejectsInvalidCSRFToken(t *testing.T) {
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
	sessionID := createSessionForUserID(t, db, "create-post-session", user.ID)

	form := url.Values{
		"title":        {"CSRF post"},
		"content":      {"This should not be created"},
		"category_ids": {"1"},
		"csrf_token":   {"forged-token"},
	}
	req := httptest.NewRequest(http.MethodPost, "/create-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	CreatePostHandler(db)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	posts, err := database.GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatalf("posts = %d, want 0", len(posts))
	}
}

// TestCreatePostHandlerRejectsInvalidSession covers unknown session tokens.
func TestCreatePostHandlerRejectsInvalidSession(t *testing.T) {
	db := openTestDB(t)

	req := httptest.NewRequest(http.MethodPost, "/create-post", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bad-session"})
	w := httptest.NewRecorder()

	CreatePostHandler(db)(w, req)

	// Invalid sessions are authentication failures.
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	// The response should identify the session parsing problem.
	if !strings.Contains(w.Body.String(), "invalid session") {
		t.Fatalf("body = %q, want invalid session message", w.Body.String())
	}
}

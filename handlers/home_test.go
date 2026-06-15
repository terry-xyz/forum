package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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

func TestHomeHandlerRendersStyledNotFoundPage(t *testing.T) {
	db := openTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Page not found") {
		t.Fatalf("body = %q, want not found message", body)
	}
	if !strings.Contains(body, "error-page") {
		t.Fatalf("body = %q, want styled error page", body)
	}
	assertSharedPageAssets(t, body)
}

// TestHomeHandlerAllowsGuestSession verifies guests can read the public feed.
func TestHomeHandlerAllowsGuestSession(t *testing.T) {
	// Guest rendering still needs the database because the feed and category
	// filters are loaded before the navigation is written.
	db := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	// Missing sessions should still render the public home page.
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Guests get auth links, while actions that submit user-owned data stay hidden.
	body := w.Body.String()
	if !strings.Contains(body, `<a href="/login">Login</a>`) {
		t.Fatalf("body = %q, want login link", body)
	}
	if strings.Contains(body, `<a href="/create-post">Create post</a>`) {
		t.Fatalf("body = %q, want create-post link hidden for guests", body)
	}
}

// TestHomeHandlerRendersFullHTMLDocument verifies home output is a complete HTML page.
func TestHomeHandlerRendersFullHTMLDocument(t *testing.T) {
	db := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := strings.TrimSpace(w.Body.String())
	if !strings.HasPrefix(body, "<!doctype html>") {
		t.Fatalf("body = %q, want document to start with doctype", body)
	}
	for _, required := range []string{"<html", "<head>", "<title>Forum</title>", "<body>", "</body>", "</html>"} {
		if !strings.Contains(body, required) {
			t.Fatalf("body = %q, want full document marker %q", body, required)
		}
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
	sessionID := createSessionForUserID(t, db, "home-session", user.ID)

	// Add a valid session cookie so authenticated navigation and forms are shown.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
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

// TestHomeHandlerRendersCommentsWithAuthors verifies comment content and authors are shown.
func TestHomeHandlerRendersCommentsWithAuthors(t *testing.T) {
	// Build a database with one author, one post, and one valid comment.
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
	if err := database.CreateComment(db, user.ID, postID, "first comment"); err != nil {
		t.Fatal(err)
	}
	sessionID := createSessionForUserID(t, db, "home-comment-session", user.ID)

	// Authenticate as the valid user so rendering includes the comment action forms.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "<h5>author</h5>") {
		t.Fatalf("body = %q, want rendered comment author", body)
	}
	if !strings.Contains(body, "<p>first comment</p>") {
		t.Fatalf("body = %q, want rendered comment content", body)
	}
}

func TestHomeHandlerRendersCommentCharacterCounterHooks(t *testing.T) {
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
	sessionID := createSessionForUserID(t, db, "home-comment-counter-session", user.ID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	required := []string{
		`data-comment-textarea`,
		`data-comment-limit="500"`,
		`data-comment-warning-threshold="450"`,
		`data-comment-counter`,
		`aria-live="polite"`,
	}
	for _, want := range required {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want comment counter hook %q", body, want)
		}
	}
}

// TestHomeHandlerLimitsDefaultFeedToFirstPage verifies the feed cannot render every stored post.
func TestHomeHandlerLimitsDefaultFeedToFirstPage(t *testing.T) {
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

	for i := 1; i <= feedPageSize+1; i++ {
		if _, err := db.Exec(
			"INSERT INTO posts (author_id, title, content, created_at) VALUES (?, ?, ?, datetime('2026-01-01', ? || ' minutes'))",
			user.ID,
			"Post "+strconv.Itoa(i),
			"Hello forum",
			i,
		); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, "<h3>Post 1</h3>") {
		t.Fatalf("body contains oldest overflow post: %q", body)
	}
	if !strings.Contains(body, "<h3>Post 21</h3>") {
		t.Fatalf("body = %q, want newest post", body)
	}
	if !strings.Contains(body, `href="/?page=2"`) {
		t.Fatalf("body = %q, want next page link", body)
	}
}

// TestHomeHandlerLimitsRenderedCommentsPerPost verifies one busy post cannot flood the feed.
func TestHomeHandlerLimitsRenderedCommentsPerPost(t *testing.T) {
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
	for i := 1; i <= feedCommentPreviewLimit+1; i++ {
		if _, err := db.Exec(
			"INSERT INTO comments (author_id, post_id, content, created_at) VALUES (?, ?, ?, datetime('2026-01-01', ? || ' minutes'))",
			user.ID,
			postID,
			"comment "+strconv.Itoa(i),
			i,
		); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, "<p>comment 1</p>") {
		t.Fatalf("body contains oldest overflow comment: %q", body)
	}
	if !strings.Contains(body, "<p>comment 6</p>") {
		t.Fatalf("body = %q, want newest comment", body)
	}
}

// TestHomeHandlerEscapesRenderedUserContent verifies submitted HTML is shown as text.
func TestHomeHandlerEscapesRenderedUserContent(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "author@example.com", `<b>author</b>`, "password"); err != nil {
		t.Fatal(err)
	}
	user, err := database.GetUserByEmail(db, "author@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}
	postID, err := database.CreatePost(db, user.ID, `<script>title</script>`, `<img src=x onerror=alert(1)>`)
	if err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", `<em>category</em>`)
	if err != nil {
		t.Fatal(err)
	}
	categoryID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AddCategoriesToPost(db, postID, []int{int(categoryID)}); err != nil {
		t.Fatal(err)
	}
	if err := database.CreateComment(db, user.ID, postID, `<strong>comment</strong>`); err != nil {
		t.Fatal(err)
	}
	sessionID := createSessionForUserID(t, db, "home-escape-session", user.ID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	HomeHandler(db)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	rawValues := []string{
		`<script>title</script>`,
		`<img src=x onerror=alert(1)>`,
		`<b>author</b>`,
		`<em>category</em>`,
		`<strong>comment</strong>`,
	}
	for _, raw := range rawValues {
		if strings.Contains(body, raw) {
			t.Fatalf("body contains unescaped user content %q: %q", raw, body)
		}
	}
	escapedValues := []string{
		`&lt;script&gt;title&lt;/script&gt;`,
		`&lt;img src=x onerror=alert(1)&gt;`,
		`&lt;b&gt;author&lt;/b&gt;`,
		`&lt;em&gt;category&lt;/em&gt;`,
		`&lt;strong&gt;comment&lt;/strong&gt;`,
	}
	for _, escaped := range escapedValues {
		if !strings.Contains(body, escaped) {
			t.Fatalf("body = %q, want escaped content %q", body, escaped)
		}
	}
}

// TestHandlerTemplatesLiveInHTMLFiles verifies rendering markup is kept in template files.
func TestHandlerTemplatesLiveInHTMLFiles(t *testing.T) {
	for _, path := range []string{
		"../templates/home.html",
		"../templates/post.html",
		"../templates/comment.html",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected handler template file %s to exist: %v", path, err)
		}
	}
}

// TestRenderPostsReportsMissingCommentAuthor preserves the render error contract for bad data.
func TestRenderPostsReportsMissingCommentAuthor(t *testing.T) {
	db := openTestDBWithoutForeignKeys(t)

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
	if _, err := db.Exec("INSERT INTO comments (author_id, post_id, content) VALUES (?, ?, ?)", 999, postID, "orphaned comment"); err != nil {
		t.Fatal(err)
	}
	posts, err := database.GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}
	_, err = buildRenderedPosts(db, posts, user, "", 0, "")

	if err == nil {
		t.Fatal("expected missing comment author to return an error")
	}
	if !strings.Contains(err.Error(), "comment author not found") {
		t.Fatalf("error = %v, want missing comment author", err)
	}
}

// openTestDBWithoutForeignKeys creates a malformed-data fixture for render error tests.
func openTestDBWithoutForeignKeys(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	schema, err := os.ReadFile("../database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}

	return db
}

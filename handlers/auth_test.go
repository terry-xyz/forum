package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"forum/database"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// TestRegisterHandlerRejectsDuplicateUser verifies duplicate account conflicts
// produce a useful client-facing response.
func TestRegisterHandlerRejectsDuplicateUser(t *testing.T) {
	// Start with an existing user so the registration insert hits a UNIQUE constraint.
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}

	// Submit the same email and username through the handler.
	form := url.Values{
		"email":    {"user@example.com"},
		"username": {"user"},
		"password": {"password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	RegisterHandler(db)(w, req)

	// Duplicate users should be reported as a conflict, not a generic 500.
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
	}

	// The body should explain that either unique field may already be taken.
	if !strings.Contains(w.Body.String(), "email or username is already taken") {
		t.Fatalf("body = %q, want duplicate-user message", w.Body.String())
	}
}

// TestRegisterHandlerRejectsInvalidRequiredFields verifies direct POSTs cannot bypass required fields.
func TestRegisterHandlerRejectsInvalidRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		user    string
		pass    string
		message string
	}{
		{
			name:    "blank email",
			email:   "   ",
			user:    "user",
			pass:    "password",
			message: "invalid email format",
		},
		{
			name:    "blank username",
			email:   "user@example.com",
			user:    "\t ",
			pass:    "password",
			message: "username cannot be empty",
		},
		{
			name:    "blank password",
			email:   "user@example.com",
			user:    "user",
			pass:    "   ",
			message: "password must be at least 8 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)

			form := url.Values{
				"email":    {tt.email},
				"username": {tt.user},
				"password": {tt.pass},
			}
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			RegisterHandler(db)(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			if !strings.Contains(w.Body.String(), tt.message) {
				t.Fatalf("body = %q, want %q", w.Body.String(), tt.message)
			}

			users, err := db.Query("SELECT id FROM users")
			if err != nil {
				t.Fatal(err)
			}
			defer users.Close()
			if users.Next() {
				t.Fatal("invalid registration inserted a user")
			}
			if err := users.Err(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestLoginHandlerRejectsUnknownEmail verifies missing accounts cannot log in.
func TestLoginHandlerRejectsUnknownEmail(t *testing.T) {
	// Use an empty database so the submitted email is guaranteed to be unknown.
	db := openTestDB(t)

	// Build a realistic form-encoded login request.
	form := url.Values{
		"email":    {"missing@example.com"},
		"password": {"password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	LoginHandler(db)(w, req)

	// Unknown users should receive the same unauthorized response as bad passwords.
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// The message intentionally avoids revealing whether the email exists.
	if !strings.Contains(w.Body.String(), "invalid email or password") {
		t.Fatalf("body = %q, want invalid credentials message", w.Body.String())
	}
}

// TestLoginHandlerRejectsWrongPassword verifies password mismatch handling.
func TestLoginHandlerRejectsWrongPassword(t *testing.T) {
	// Insert a real account so the failure is specifically the password check.
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", hashPasswordForTest(t, "password")); err != nil {
		t.Fatal(err)
	}

	// Submit the known email with an incorrect password.
	form := url.Values{
		"email":    {"user@example.com"},
		"password": {"wrong-password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	LoginHandler(db)(w, req)

	// Bad credentials should not set a session or redirect.
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// The response text should match the unknown-email branch.
	if !strings.Contains(w.Body.String(), "invalid email or password") {
		t.Fatalf("body = %q, want invalid credentials message", w.Body.String())
	}
}

// TestLoginHandlerSetsSessionCookie verifies successful login creates a cookie.
func TestLoginHandlerSetsSessionCookie(t *testing.T) {
	// Insert the account that the login request will authenticate.
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", hashPasswordForTest(t, "password")); err != nil {
		t.Fatal(err)
	}

	// Submit matching credentials as form data.
	form := url.Values{
		"email":    {"user@example.com"},
		"password": {"password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	LoginHandler(db)(w, req)

	// Successful login redirects to the home page.
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	// The handler should set exactly one session cookie in this response.
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}

	// Validate the cookie attributes that matter for later requests and browser security.
	cookie := cookies[0]
	if cookie.Name != "session" {
		t.Fatalf("cookie name = %q, want session", cookie.Name)
	}
	if cookie.Value == "" {
		t.Fatal("cookie value is empty")
	}
	if !cookie.HttpOnly {
		t.Fatal("session cookie should be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("same site = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
	if cookie.MaxAge <= 0 {
		t.Fatalf("max age = %d, want positive value", cookie.MaxAge)
	}

	user, err := database.GetUserByEmail(db, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := database.GetUserIDBySessionID(db, cookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	if userID != user.ID {
		t.Fatalf("session user id = %d, want %d", userID, user.ID)
	}
}

// TestLoginHandlerInvalidatesPreviousSessions verifies one active session per user.
func TestLoginHandlerInvalidatesPreviousSessions(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", hashPasswordForTest(t, "password")); err != nil {
		t.Fatal(err)
	}
	user, err := database.GetUserByEmail(db, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}

	oldSessionID := createSessionForUserID(t, db, "old-session", user.ID)

	form := url.Values{
		"email":    {"user@example.com"},
		"password": {"password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	LoginHandler(db)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	newSessionID := cookies[0].Value
	if newSessionID == "" {
		t.Fatal("new session cookie value is empty")
	}

	oldUserID, err := database.GetUserIDBySessionID(db, oldSessionID)
	if err != nil {
		t.Fatal(err)
	}
	if oldUserID != 0 {
		t.Fatalf("old session user id = %d, want 0", oldUserID)
	}

	newUserID, err := database.GetUserIDBySessionID(db, newSessionID)
	if err != nil {
		t.Fatal(err)
	}
	if newUserID != user.ID {
		t.Fatalf("new session user id = %d, want %d", newUserID, user.ID)
	}
}

// TestLogoutHandlerExpiresSessionCookie verifies logout clears browser session state.
func TestLogoutHandlerExpiresSessionCookie(t *testing.T) {
	db := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()

	LogoutHandler(db)(w, req)

	// Logout redirects after writing the expired cookie.
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	// The response should contain the replacement session cookie.
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}

	// The cookie name and security attributes must match the login cookie, while
	// MaxAge must tell the browser to remove it.
	cookie := cookies[0]
	if cookie.Name != "session" {
		t.Fatalf("cookie name = %q, want session", cookie.Name)
	}
	if cookie.MaxAge >= 0 {
		t.Fatalf("max age = %d, want expired cookie", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Fatal("session cookie should be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("same site = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
}

// TestLogoutHandlerDeletesServerSession verifies logout invalidates the stored token.
func TestLogoutHandlerDeletesServerSession(t *testing.T) {
	db := openTestDB(t)
	sessionID := createSessionForTest(t, db, "user@example.com")

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	LogoutHandler(db)(w, req)

	userID, err := database.GetUserIDBySessionID(db, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if userID != 0 {
		t.Fatalf("session user id after logout = %d, want 0", userID)
	}
}

// openTestDB creates a temporary handlers test database from the shared schema.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Each test gets an isolated SQLite file under the test temp directory.
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	// Handler tests live one directory below database/schema.sql.
	schema, err := os.ReadFile("../database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Apply the full schema before the handler under test runs.
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}

	return db
}

// hashPasswordForTest returns a bcrypt hash so login tests match stored production credentials.
func hashPasswordForTest(t *testing.T, password string) string {
	t.Helper()

	// LoginHandler compares submitted passwords against bcrypt hashes, so test
	// fixtures must store the same format that RegisterHandler writes.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	return string(hashedPassword)
}

// createSessionForTest creates a user and maps a stable session token to it.
func createSessionForTest(t *testing.T, db *sql.DB, email string) string {
	t.Helper()

	if err := database.CreateUser(db, email, "user-"+email, hashPasswordForTest(t, "password")); err != nil {
		t.Fatal(err)
	}
	user, err := database.GetUserByEmail(db, email)
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}

	return createSessionForUserID(t, db, "session-"+email, user.ID)
}

// createSessionForUserID maps a stable session token to an existing test user.
func createSessionForUserID(t *testing.T, db *sql.DB, sessionID string, userID int) string {
	t.Helper()

	expiresAt := time.Now().Add(time.Hour)
	if err := database.CreateSession(db, sessionID, userID, expiresAt); err != nil {
		t.Fatal(err)
	}

	return sessionID
}

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

// TestRegisterHandlerDoesNotRevealDuplicateUser verifies duplicate account
// conflicts do not expose whether an email or username already exists.
func TestRegisterHandlerDoesNotRevealDuplicateUser(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}

	newForm := url.Values{
		"email":    {"new@example.com"},
		"username": {"new-user"},
		"password": {"password"},
	}
	newReq := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(newForm.Encode()))
	newReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	newResponse := httptest.NewRecorder()

	RegisterHandler(db)(newResponse, newReq)

	duplicateForm := url.Values{
		"email":    {"user@example.com"},
		"username": {"user"},
		"password": {"password"},
	}
	duplicateReq := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(duplicateForm.Encode()))
	duplicateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	duplicateResponse := httptest.NewRecorder()

	RegisterHandler(db)(duplicateResponse, duplicateReq)

	if duplicateResponse.Code != newResponse.Code {
		t.Fatalf("duplicate status = %d, want %d", duplicateResponse.Code, newResponse.Code)
	}
	if duplicateResponse.Header().Get("Location") != newResponse.Header().Get("Location") {
		t.Fatalf("duplicate location = %q, want %q", duplicateResponse.Header().Get("Location"), newResponse.Header().Get("Location"))
	}
	if duplicateResponse.Body.String() != newResponse.Body.String() {
		t.Fatalf("duplicate body = %q, want %q", duplicateResponse.Body.String(), newResponse.Body.String())
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
	if cookie.Secure {
		t.Fatal("local http session cookie should not be Secure")
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

// TestLoginHandlerSetsSecureCookieForHTTPS verifies HTTPS logins get Secure cookies.
func TestLoginHandlerSetsSecureCookieForHTTPS(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", hashPasswordForTest(t, "password")); err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"email":    {"user@example.com"},
		"password": {"password"},
	}
	req := httptest.NewRequest(http.MethodPost, "https://example.com/login", strings.NewReader(form.Encode()))
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
	if !cookies[0].Secure {
		t.Fatal("https session cookie should be Secure")
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

// TestLogoutHandlerRejectsGET verifies logout cannot be triggered by top-level navigation.
func TestLogoutHandlerRejectsGET(t *testing.T) {
	db := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()

	LogoutHandler(db)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// TestLogoutHandlerRejectsMissingCSRFToken verifies a session cookie alone cannot log out.
func TestLogoutHandlerRejectsMissingCSRFToken(t *testing.T) {
	db := openTestDB(t)
	sessionID := createSessionForTest(t, db, "user@example.com")

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	LogoutHandler(db)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	userID, err := database.GetUserIDBySessionID(db, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if userID == 0 {
		t.Fatal("session was deleted without a csrf token")
	}
}

// TestLogoutHandlerExpiresSessionCookie verifies logout clears browser session state.
func TestLogoutHandlerExpiresSessionCookie(t *testing.T) {
	db := openTestDB(t)
	sessionID := createSessionForTest(t, db, "user@example.com")
	csrfToken := csrfTokenForTest(t, db, sessionID)

	form := url.Values{"csrf_token": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
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
	if cookie.Secure {
		t.Fatal("local http session-clearing cookie should not be Secure")
	}
}

// TestLogoutHandlerSetsSecureCookieBehindHTTPSProxy verifies proxied HTTPS clears Secure cookies.
func TestLogoutHandlerSetsSecureCookieBehindHTTPSProxy(t *testing.T) {
	db := openTestDB(t)
	sessionID := createSessionForTest(t, db, "user@example.com")
	csrfToken := csrfTokenForTest(t, db, sessionID)

	form := url.Values{"csrf_token": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()

	LogoutHandler(db)(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	if !cookies[0].Secure {
		t.Fatal("proxied https session-clearing cookie should be Secure")
	}
}

// TestLogoutHandlerDeletesServerSession verifies logout invalidates the stored token.
func TestLogoutHandlerDeletesServerSession(t *testing.T) {
	db := openTestDB(t)
	sessionID := createSessionForTest(t, db, "user@example.com")
	csrfToken := csrfTokenForTest(t, db, sessionID)

	form := url.Values{"csrf_token": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db?_foreign_keys=on")
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
	if err := database.CreateSession(db, sessionID, userID, "csrf-"+sessionID, expiresAt); err != nil {
		t.Fatal(err)
	}

	return sessionID
}

// csrfTokenForTest returns the server-side token tied to a test session.
func csrfTokenForTest(t *testing.T, db *sql.DB, sessionID string) string {
	t.Helper()

	token, err := database.GetCSRFTokenBySessionID(db, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("csrf token is empty")
	}

	return token
}

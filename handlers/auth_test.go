package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"forum/database"

	_ "github.com/mattn/go-sqlite3"
)

func TestRegisterHandlerRejectsDuplicateUser(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"email":    {"user@example.com"},
		"username": {"user"},
		"password": {"password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	RegisterHandler(db)(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
	}

	if !strings.Contains(w.Body.String(), "email or username is already taken") {
		t.Fatalf("body = %q, want duplicate-user message", w.Body.String())
	}
}

func TestLoginHandlerRejectsUnknownEmail(t *testing.T) {
	db := openTestDB(t)

	form := url.Values{
		"email":    {"missing@example.com"},
		"password": {"password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	LoginHandler(db)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	if !strings.Contains(w.Body.String(), "invalid email or password") {
		t.Fatalf("body = %q, want invalid credentials message", w.Body.String())
	}
}

func TestLoginHandlerRejectsWrongPassword(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"email":    {"user@example.com"},
		"password": {"wrong-password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	LoginHandler(db)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	if !strings.Contains(w.Body.String(), "invalid email or password") {
		t.Fatalf("body = %q, want invalid credentials message", w.Body.String())
	}
}

func TestLoginHandlerSetsSessionCookie(t *testing.T) {
	db := openTestDB(t)

	if err := database.CreateUser(db, "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}

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
}

func TestLogoutHandlerExpiresSessionCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()

	LogoutHandler()(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}

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

func openTestDB(t *testing.T) *sql.DB {
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

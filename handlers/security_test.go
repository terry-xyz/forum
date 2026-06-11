package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseLimitedFormRejectsUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register?email=user@example.com&username=user&password=password", strings.NewReader("ignored"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	if parseLimitedForm(w, req) {
		t.Fatal("parseLimitedForm accepted a POST with text/plain content")
	}
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnsupportedMediaType)
	}
	assertStyledErrorPage(t, w.Body.String(), "unsupported content type")
}

func TestValidCSRFTokenRejectsTokenProvidedOnlyInQuery(t *testing.T) {
	db := openTestDB(t)
	sessionID := createSessionForTest(t, db, "query-csrf@example.com")
	csrfToken := csrfTokenForTest(t, db, sessionID)

	body := url.Values{"title": {"post title"}}
	req := httptest.NewRequest(http.MethodPost, "/posts?csrf_token="+url.QueryEscape(csrfToken), strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if !parseLimitedForm(w, req) {
		t.Fatalf("parseLimitedForm rejected form: status = %d, body = %q", w.Code, w.Body.String())
	}
	if validCSRFToken(db, req, sessionID) {
		t.Fatal("validCSRFToken accepted csrf_token from the URL query")
	}
}

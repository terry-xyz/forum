package handlers

import (
	"net/http"
	"net/http/httptest"
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
}

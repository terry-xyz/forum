package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandlerRejectsUnsupportedMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(nil)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	if !strings.Contains(w.Body.String(), "method not allowed") {
		t.Fatalf("body = %q, want method-not-allowed message", w.Body.String())
	}
}

func TestHomeHandlerRejectsMissingSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(nil)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	if !strings.Contains(w.Body.String(), "unauthorized") {
		t.Fatalf("body = %q, want unauthorized message", w.Body.String())
	}
}

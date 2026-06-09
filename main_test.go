package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPServerConfiguresTimeouts(t *testing.T) {
	server := newHTTPServer(":0")

	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s, want 5s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("ReadTimeout = %s, want 10s", server.ReadTimeout)
	}
	if server.WriteTimeout != 15*time.Second {
		t.Fatalf("WriteTimeout = %s, want 15s", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want 60s", server.IdleTimeout)
	}
}

func TestNewHTTPServerAddsSecurityHeaders(t *testing.T) {
	server := newHTTPServer(":0")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.Handler.ServeHTTP(w, req)

	headers := w.Result().Header
	wantHeaders := map[string]string{
		"Content-Security-Policy": "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; object-src 'none'; upgrade-insecure-requests",
		"X-Frame-Options":         "DENY",
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
	}
	for name, want := range wantHeaders {
		if got := headers.Get(name); got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

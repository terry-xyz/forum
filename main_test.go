package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConfiguredServerAddressDefaultsToLocalPort(t *testing.T) {
	t.Setenv("PORT", "")

	if got := configuredServerAddress(); got != ":8080" {
		t.Fatalf("configuredServerAddress() = %q, want %q", got, ":8080")
	}
}

func TestConfiguredServerAddressUsesPortEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")

	if got := configuredServerAddress(); got != ":9090" {
		t.Fatalf("configuredServerAddress() = %q, want %q", got, ":9090")
	}
}

func TestConfiguredServerAddressAcceptsFullAddress(t *testing.T) {
	t.Setenv("PORT", "127.0.0.1:9090")

	if got := configuredServerAddress(); got != "127.0.0.1:9090" {
		t.Fatalf("configuredServerAddress() = %q, want %q", got, "127.0.0.1:9090")
	}
}

func TestConfiguredDatabasePathDefaultsToForumDB(t *testing.T) {
	t.Setenv("DATABASE_PATH", "")

	if got := configuredDatabasePath(); got != "forum.db" {
		t.Fatalf("configuredDatabasePath() = %q, want %q", got, "forum.db")
	}
}

func TestConfiguredDatabasePathUsesEnvironment(t *testing.T) {
	t.Setenv("DATABASE_PATH", "data/staging.db")

	if got := configuredDatabasePath(); got != "data/staging.db" {
		t.Fatalf("configuredDatabasePath() = %q, want %q", got, "data/staging.db")
	}
}

func TestNewHTTPServerConfiguresTimeouts(t *testing.T) {
	server := newHTTPServer(":0", http.NewServeMux())

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
	server := newHTTPServer(":0", http.NewServeMux())

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

func TestNewHTTPServerServesStaticAssets(t *testing.T) {
	mux := newAppMux()
	server := newHTTPServer(":0", mux)

	tests := []struct {
		path        string
		contentType string
	}{
		{path: "/static/style.css", contentType: "text/css"},
		{path: "/static/theme.js", contentType: "text/javascript"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		w := httptest.NewRecorder()

		server.Handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", tt.path, w.Code, http.StatusOK)
		}
		if got := w.Result().Header.Get("Content-Type"); !strings.Contains(got, tt.contentType) {
			t.Fatalf("GET %s Content-Type = %q, want %s content type", tt.path, got, tt.contentType)
		}
		if got := w.Result().Header.Get("Content-Security-Policy"); got != contentSecurityPolicy {
			t.Fatalf("GET %s CSP = %q, want %q", tt.path, got, contentSecurityPolicy)
		}
	}
}

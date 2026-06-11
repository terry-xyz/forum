package handlers

import (
	"strings"
	"testing"
)

func assertSharedPageAssets(t *testing.T, body string) {
	t.Helper()

	required := []string{
		"<!doctype html>",
		`<html lang="en">`,
		`<script src="/static/theme.js"></script>`,
		`<link rel="stylesheet" href="/static/style.css">`,
		`data-theme-toggle`,
		"</html>",
	}
	for _, want := range required {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want shared page asset %q", body, want)
		}
	}
}

func assertStyledErrorPage(t *testing.T, body string, message string) {
	t.Helper()

	assertSharedPageAssets(t, body)
	if !strings.Contains(body, "error-page") {
		t.Fatalf("body = %q, want styled error page", body)
	}
	if !strings.Contains(body, message) {
		t.Fatalf("body = %q, want error message %q", body, message)
	}
}

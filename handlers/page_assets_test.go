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

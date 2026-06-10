package handlers

import (
	"html"
	"net/http"
)

func renderFormPage(w http.ResponseWriter, title string, formHTML string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>` + html.EscapeString(title) + ` - Forum</title>
	<script src="/static/theme.js"></script>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<nav class="site-nav" aria-label="Primary">
		<div class="category-nav" aria-label="Categories">
			<a href="/">All posts</a>
		</div>
		<div class="account-nav">
			<a href="/login">Login</a>
			<a href="/register">Register</a>
			<button type="button" class="theme-toggle" data-theme-toggle aria-label="Switch color theme" aria-pressed="false">
				<span data-theme-toggle-label>Theme</span>
			</button>
		</div>
	</nav>
	<main class="page-shell form-page">
		<section class="form-card">
			<h1>` + html.EscapeString(title) + `</h1>
` + formHTML + `
		</section>
	</main>
</body>
</html>`))
}

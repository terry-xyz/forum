# Forum Responsive Theme Design

## Goal

Add maintainable styling to the forum so it works across screen sizes, uses smooth but restrained animations, follows the user's device light/dark preference by default, and provides a manual theme switch.

## Current State

- `templates/home.html` is the full page shell and links `/static/style.css`.
- `templates/post.html` and `templates/comment.html` render repeated content fragments.
- `static/style.css` exists but is not yet populated.
- The Go server does not currently register a `/static/` file handler, so static assets must be served before CSS or JavaScript can load.
- The Content Security Policy allows same-origin external assets, so an external `/static/theme.js` file fits the current security model better than inline JavaScript.

## Approach

Use small, explicit changes:

- Register `/static/` with `http.FileServer` in `main.go`.
- Add semantic classes and a theme toggle button to `templates/home.html`.
- Keep `post.html` and `comment.html` changes minimal, adding only classes needed for reliable styling.
- Implement theme variables in `static/style.css` using `prefers-color-scheme` as the default.
- Implement manual theme override in `static/theme.js` using `localStorage` and a `data-theme` attribute on `<html>`.

## User Experience

The page should feel like a practical forum interface, not a landing page. The content remains the primary focus. On desktop, the forum uses a readable centered layout with a flexible top navigation. On mobile, navigation and forms wrap cleanly, inputs fill available width, and post actions remain usable without horizontal scrolling.

The theme button appears in the top navigation. Initial theme follows the device preference. When the user toggles theme, the choice persists locally. Motion uses short transitions for hover, focus, button presses, and content appearance, and disables non-essential animation for users who prefer reduced motion.

## Components

- Static server: exposes `/static/style.css` and `/static/theme.js`.
- Page shell: owns the stylesheet link, theme script, and top navigation theme button.
- CSS theme system: central custom properties for color, shadow, borders, spacing, and motion.
- Theme script: reads the stored preference, applies it before user interaction, updates the button state, and toggles between light and dark.

## Error Handling

If `localStorage` is unavailable, the page still uses the system preference and the toggle works for the current page load. If JavaScript fails, CSS still provides responsive layout and automatic light/dark behavior through `prefers-color-scheme`.

## Verification

- Run `go test ./...`.
- Start the server and confirm `/`, `/static/style.css`, and `/static/theme.js` return successful responses.
- Manually inspect desktop and mobile widths for wrapping, readability, theme switching, and no horizontal overflow.

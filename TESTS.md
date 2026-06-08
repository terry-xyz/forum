# Basic forum status

This file tracks only the basic forum requirements. Extras are intentionally out
of scope.

## Verified

- `go test ./...` passes.
- Test files exist for handlers, database helpers, schema execution, seed data,
  and session ID generation.
- Docker image builds successfully, and the container serves the forum on port
  `8080`.
- SQLite is used, and the schema contains `CREATE`, `INSERT`, and application
  code uses `SELECT`.
- Registration asks for email, username, and password.
- Registration rejects invalid/empty email, empty username, and empty/short
  password submissions with `400 Bad Request`.
- Duplicate email/username registration is rejected.
- Login rejects unknown emails and wrong passwords.
- Successful login creates an expiring `session` cookie with `HttpOnly` and
  `SameSite=Lax`.
- Successful login invalidates the user's previous sessions, so only the newest
  browser session remains active.
- Logout deletes the server session and expires the browser cookie.
- Guests can view the home page but do not see create/comment/reaction forms.
- Registered users can create posts with one or more categories.
- Empty or whitespace-only posts are rejected with `400 Bad Request`.
- Empty comments are rejected with `400 Bad Request`.
- Posts can be filtered by category, by current user's created posts, and by
  current user's liked posts.
- Post/comment reactions are stored as one reaction per user per target, so a
  user cannot like and dislike the same post/comment at the same time.
- User-generated post titles, post content, usernames, category names, and
  comment content are rendered through file-backed `html/template` templates and
  escaped in the feed.
- Passwords are stored with bcrypt hashes.
- The users table requires a stored password value.
- Foreign keys prevent orphan posts, comments, category links, sessions, and
  reactions when SQLite foreign-key enforcement is enabled.
- `main.go` closes the database handle when the server exits.

## Remaining basic work

None for the basic forum audit items tracked here.

## Useful cleanup

- Consider a small session/current-user helper to reduce repeated cookie lookup
  and session validation in handlers.
- Consider query optimizations for rendering posts. `buildRenderedPosts` currently does
  several per-post/per-comment lookups for authors, categories, comments, and
  reaction counts. This is not an immediate audit blocker, but it is worth
  improving after the basic gaps are closed.

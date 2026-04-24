[ ] Create the minimal runnable Go forum skeleton: `go.mod`, server entrypoint, HTTP router, server-rendered template/layout setup, config/env loading, and baseline `GET /` plus error plumbing.
[ ] Add the primary Docker delivery path immediately: `Dockerfile`, container startup wiring, writable SQLite path/config, and volume-backed persistence so `docker build` and `docker run` work before feature work expands.
[ ] Bootstrap SQLite with `PRAGMA foreign_keys=ON`, migrations/startup schema creation, and idempotent category seeding for `users`, `sessions`, `categories`, `posts`, `post_categories`, `comments`, and `reactions`.
[ ] Enforce schema constraints and delete behavior in persistence: unique `users.email`, unique `users.username`, unique `sessions.token`, one active reaction per user/target, and cascades for post/comment deletion.
[ ] Implement the repository/service layer with transaction boundaries for registration, session replacement, reaction switching, and author deletes.
[ ] Implement registration on `GET /register` and `POST /register` with required `email`, `username`, and `password`, plus independent duplicate-email and duplicate-username rejection with no partial account creation.
[ ] Implement login/logout on `GET /login`, `POST /login`, and `POST /logout` using `email + password`, secure server-side sessions, one active session per user, and supersession of older sessions.
[ ] Implement auth middleware for protected reads/writes that detects invalid, expired, revoked, and superseded sessions, clears cookies, shows human-readable re-login messaging, and preserves then consumes return paths after successful login.
[ ] Implement the public homepage feed on `GET /` with newest-first ordering and URL-reconstructible strict-intersection filters across selected categories, `my posts`, and positive-only `liked posts`, including empty-state reset guidance.
[ ] Implement authenticated post creation on `POST /posts` with title/body validation, one-or-more seeded categories, author display-name snapshotting, and PRG redirect behavior.
[ ] Implement public post detail on `GET /posts/{id}` with full post content, categories, author snapshots, reaction totals, and flat oldest-first comments.
[ ] Implement comment creation on `POST /posts/{id}/comments` with validation, author display-name snapshotting, and PRG redirect back to the post detail page.
[ ] Implement reactions on `POST /reactions` for posts and comments with auth enforcement, one active reaction per user/target, and atomic like/dislike switching.
[ ] Implement author-only hard deletes on `POST /posts/{id}/delete` and `POST /comments/{id}/delete` so cascaded removals disappear from feed, detail, and filtered views.
[ ] Add consistent request/error handling for malformed writes (`400`), missing routes/resources/posts/comments (`404`), and generic unexpected failures (`500`).
[ ] Add Go test coverage for auth/session lifecycle, return-path handling, filter semantics, positive-only liked posts, PRG flows, oldest-first comments, reactions, cascades, error paths, Docker build/run, and SQLite durability across container replacement.

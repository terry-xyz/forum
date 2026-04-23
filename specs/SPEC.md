# Forum Spec Plan

## Summary
Build a Go + SQLite web forum with server-rendered pages as the main implementation style, but allow richer JavaScript-driven feed behavior where it materially improves UX. The core product supports registration/login, posts with fixed categories, flat comments, likes/dislikes on posts and comments, strict filter intersection, author-only deletion, and Docker as the primary delivery artifact.

## Product Goals
- Support anonymous reading and authenticated participation.
- Keep the interface content-first, readable, and restrained.
- Use SQLite for all persistent state.
- Use cookies plus server-side session records for authentication.
- Make filter state reconstructible from URL query parameters alone.
- Ship a Docker-based run path as part of the core deliverable.

## Core Functional Scope

### Authentication
- Registration requires `email`, `username`, and `password`.
- `email` must be unique.
- Passwords are stored hashed with `bcrypt`.
- Login creates a server-side session record stored in SQLite.
- Session identity is carried in a cookie.
- Each user may have only one active session at a time.
- Creating a new session revokes any previous active session for that user.
- Expired or revoked sessions must clear the cookie and redirect to login with a human-readable explanation.
- Protected actions must redirect to login and preserve a return path.

### Content Model
- Registered users can create posts.
- Registered users can create comments on posts.
- Posts can belong to one or more categories.
- Categories come from a fixed seeded set in the database.
- Comments are flat only, with no nested replies in v1.
- Comments are rendered oldest-first on the post detail page.
- Posts and comments store an author display-name snapshot at creation time so historical authorship remains visible if account data changes later.
- Editing is out of scope in v1.
- Deletion is author-only and implemented as hard delete.
- Deleting a post cascades to its comments, category links, and reactions.
- Deleting a comment cascades to its reactions.

### Reactions
- Registered users can like or dislike posts.
- Registered users can like or dislike comments.
- A user may have at most one active reaction for a given target.
- Switching from like to dislike, or dislike to like, replaces the previous reaction atomically.
- Duplicate rapid requests must collapse safely through database constraints and idempotent handling.
- Reaction totals are visible to all users.

### Browsing and Filtering
- The homepage shows the feed newest-first by default.
- Feed filters are driven by URL query parameters.
- Supported filter dimensions:
  - category selection
  - my posts
  - liked posts
- Filter behavior is strict intersection across all active criteria.
- Empty filtered states must explain why no results matched and provide a reset path.
- Feed state must be restorable from the URL alone.
- Server-rendered fallback must remain usable even if richer JavaScript feed enhancements are added.

### UX and Error Handling
- Use Post/Redirect/Get for mutating form submissions.
- Duplicate submissions should be handled idempotently where practical.
- Unexpected server failures must return correct HTTP status codes and friendly generic error pages.
- Detailed failure diagnostics stay in server logs, not in browser output.
- Missing resources should return `404`.
- Malformed requests should return `400` where applicable.

### Delivery
- Docker is part of the main delivery path, not an optional extra.
- `docker build` and `docker run` must work against the application with SQLite persistence.
- A local non-Docker run path may also exist for development.

## Data Model

### Required Tables
- `users`
- `sessions`
- `categories`
- `posts`
- `post_categories`
- `comments`
- `reactions`

### Required Constraints
- `users.email` unique
- `sessions.token` unique
- `reactions (user_id, target_type, target_id)` unique

### Data Notes
- `users` stores identity and password hash.
- `sessions` stores token, user id, expiry, and revocation state.
- `categories` stores a fixed seeded set.
- `posts` stores title, body, author id, author display snapshot, timestamps.
- `post_categories` implements many-to-many post/category links.
- `comments` stores body, post id, author id, author display snapshot, timestamps.
- `reactions` stores user id, target type, target id, reaction value, timestamps.

## Route Surface

### Read Routes
- `GET /` for feed and filters
- `GET /posts/{id}` for post detail and flat comments
- `GET /register`
- `GET /login`

### Write Routes
- `POST /register`
- `POST /login`
- `POST /logout`
- `POST /posts`
- `POST /posts/{id}/comments`
- `POST /reactions`
- `POST /posts/{id}/delete`
- `POST /comments/{id}/delete`

### Route Behavior Requirements
- Anonymous users can read feed and post detail.
- Protected routes redirect unauthenticated users to login with a return path.
- After successful writes, handlers redirect rather than render directly.

## Implementation Decisions

### Rendering Strategy
- Prefer server-rendered HTML templates for all primary views.
- Keep templates simple and content-focused.
- If JavaScript feed enhancements are added, they must preserve usable non-JS navigation.

### Session Strategy
- Use secure random session tokens.
- Persist sessions in SQLite with expiry timestamps.
- On session lookup:
  - invalid token -> clear cookie
  - revoked session -> clear cookie
  - expired session -> clear cookie
- In all three cases, redirect the user to login when the route requires authentication.

### Filtering Semantics
- Category filters match posts tagged with all selected categories only if multiple categories are active.
- `my posts` limits to posts authored by the current user.
- `liked posts` limits to posts positively or negatively reacted to by the current user, depending on implementation choice, but behavior must be clearly defined and consistent.
- Combining filters means all active conditions must match simultaneously.

### Deletion Semantics
- Only the original author may delete their own post or comment.
- Deletion is irreversible in v1.
- Hard deletes rely on foreign keys and cascade constraints where possible.

## Test Plan

### Authentication
- register success
- duplicate email rejected
- login success
- login failure on wrong credentials
- revoked session redirects and clears cookie
- expired session redirects and clears cookie
- one-session-per-user behavior revokes older session

### Posting and Comments
- authenticated post creation succeeds
- anonymous post creation redirects to login
- authenticated comment creation succeeds
- anonymous comment creation redirects to login
- empty post rejected with `400` or form error flow
- empty comment rejected with `400` or form error flow
- flat comments render oldest-first
- author-only deletion enforced
- deleting a post removes dependent comments and reactions
- deleting a comment removes dependent reactions

### Reactions
- like/dislike requires authentication
- like on post succeeds
- dislike on post succeeds
- like/dislike on comment succeeds
- switching reaction replaces prior reaction
- repeated rapid submissions do not create multiple active rows

### Filtering and Feed
- default feed is newest-first
- category filter works
- my posts works for logged-in user
- liked posts works for logged-in user
- strict intersection yields correct result set
- empty filtered state explains active filters and provides reset
- URL query parameters fully restore view state

### Error Handling
- `404` for missing resources
- `400` for malformed input where applicable
- `500` renders a generic error page without internal details

### Delivery
- Go test suite passes
- app binary builds successfully
- Docker image builds successfully
- Docker container runs the app successfully against SQLite

## Non-Goals
- No admin role
- No moderation console
- No user-created categories
- No nested comments
- No editing of posts or comments
- No edit history

## Assumptions
- The initial category list is fixed and seeded on startup or migration.
- SQLite foreign keys are enabled explicitly.
- The application can store its SQLite database in a writable local path inside and outside Docker.
- A basic JavaScript enhancement layer is optional, but core reading and writing flows must remain fully functional without it.

# Forum

A small server-rendered forum application written in Go. It uses `net/http`,
SQLite, bcrypt password hashing, and HTML templates to provide basic forum
workflows without a separate frontend build step.

## Features

- User registration, login, logout, and expiring session cookies
- Public forum feed with category filtering
- Authenticated post and comment creation
- Like/dislike reactions for posts and comments
- User-specific views for created posts and liked posts
- Owner-only post and comment deletion
- SQLite schema initialization on startup
- Optional deterministic seed data for local testing

## Requirements

- Go `1.24` or newer
- A C toolchain with CGO enabled for `github.com/mattn/go-sqlite3`
- Docker, optional
- `make`, optional

## Quick Start

```sh
go run .
```

Open `http://localhost:8080`.

The app creates `forum.db` in the repository root and applies
`database/schema.sql` automatically on startup.

## Seed Data

Populate the local database with sample users, posts, comments, and reactions:

```sh
go run ./cmd/seed
```

Seeded accounts use `password123` as the password.

## Make Targets

```sh
make run              # run the server
make test             # run all tests
make docker-build     # build the Docker image
make docker-run       # run the container on port 8080
make docker-run-persist
```

`docker-run-persist` mounts `forum.db` into the container so data survives
between container runs.

## Docker

Build and run the app:

```sh
docker build -t forum .
docker run --rm -p 8080:8080 forum
```

For persistent SQLite data:

```sh
docker run --rm -p 8080:8080 -v "$(pwd)/forum.db:/app/forum.db" forum
```

## Testing

```sh
go test ./...
```

The test suite covers database schema setup, seed data, session ID generation,
authentication handlers, post handlers, and home page rendering behavior.

## Project Layout

```text
cmd/seed/      Seed command for local demo data
database/      SQLite schema, connection setup, and query helpers
handlers/      HTTP route handlers
helpers/       Shared helper functions
models/        Data models
templates/     HTML templates and template loader
main.go        Application entrypoint
```

## License

MIT. See `LICENSE`.

# Forum

[![CI](https://github.com/terry-xyz/forum/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/terry-xyz/forum/actions/workflows/ci.yml)
[![Container](https://github.com/terry-xyz/forum/actions/workflows/container.yml/badge.svg?branch=main)](https://github.com/terry-xyz/forum/actions/workflows/container.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/terry-xyz/forum)](https://github.com/terry-xyz/forum/blob/main/go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/terry-xyz/forum)](https://goreportcard.com/report/github.com/terry-xyz/forum)
[![License](https://img.shields.io/github/license/terry-xyz/forum)](https://github.com/terry-xyz/forum/blob/main/LICENSE)

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
- Global browser security headers for CSP, frame protection, MIME sniffing, and referrers
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

To use a different port or SQLite file:

```sh
PORT=9090 DATABASE_PATH=data/forum.db go run .
```

## Seed Data

Populate the local database with sample users, posts, comments, and reactions:

```sh
go run ./cmd/seed
```

Seeded accounts use `password123` as the password.

## Make Targets

```sh
make run              # run the server
make run-config       # run with PORT and DATABASE_PATH from this Makefile
make seed             # populate the configured database
make test             # run all tests
make test-long        # repeat the full test suite
make docker-build     # build the Docker image
make docker-rebuild   # rebuild the Docker image without cache
make docker-test      # run tests inside the Go Docker image
make docker-test-long # repeat Docker tests
make docker-run       # run the container on port 8080
make docker-run-persist
make docker-run-detached
make docker-logs
make docker-stop
```

Set `PORT` to change the host port and `DATABASE_PATH` to change the local
database path. `docker-run-persist` and `docker-run-detached` mount `data/`
into the container and store SQLite data at `/app/data/forum.db`.

## Docker

The Docker image runs the forum process as non-root UID `10001` and pins base
images by digest for reproducible rebuilds.

Build and run the app:

```sh
docker build -t forum .
docker run --rm -p 8080:8080 forum
```

For persistent SQLite data:

```sh
mkdir -p data
docker run --rm -p 8080:8080 -e DATABASE_PATH=/app/data/forum.db -v "$(pwd)/data:/app/data" forum
```

If your host enforces Unix file ownership on bind mounts, make sure the mounted
data directory is writable by UID `10001`.

## Deployment

GitHub Actions publishes the Docker image to GitHub Container Registry for
pushes to `main` and for version tags. Pull the latest image and persist the
SQLite database with a mounted data directory:

```sh
docker pull ghcr.io/terry-xyz/forum:latest
mkdir -p data
docker run -d --name forum --restart unless-stopped \
  -p 8080:8080 \
  -e PORT=8080 \
  -e DATABASE_PATH=/app/data/forum.db \
  -v "$(pwd)/data:/app/data" \
  ghcr.io/terry-xyz/forum:latest
```

Version tags are published with the same tag, for example
`ghcr.io/terry-xyz/forum:0.4.0`. The image package may need to be made public
in the repository's GitHub Packages settings before unauthenticated pulls.

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

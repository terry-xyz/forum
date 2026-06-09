package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestSchemaExecutes verifies that the SQL schema can initialize a fresh DB.
func TestSchemaExecutes(t *testing.T) {
	// Use a temporary database file so the test never touches developer data.
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Read the package-local schema file exactly as application setup does.
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Executing the full schema should succeed on an empty SQLite database.
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("schema did not execute: %v", err)
	}
}

// TestUsersPasswordIsRequired verifies the schema rejects users without stored passwords.
func TestUsersPasswordIsRequired(t *testing.T) {
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("schema did not execute: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (email, username) VALUES (?, ?)", "user@example.com", "user")
	if err == nil {
		t.Fatal("expected missing password insert to fail")
	}
}

// TestSchemaRejectsOrphanedRelationships verifies related rows cannot point at missing parents.
func TestSchemaRejectsOrphanedRelationships(t *testing.T) {
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("schema did not execute: %v", err)
	}

	tests := []struct {
		name  string
		query string
		args  []any
	}{
		{
			name:  "post author",
			query: "INSERT INTO posts (author_id, title, content) VALUES (?, ?, ?)",
			args:  []any{999, "orphaned post", "missing author"},
		},
		{
			name:  "comment author and post",
			query: "INSERT INTO comments (author_id, post_id, content) VALUES (?, ?, ?)",
			args:  []any{999, 999, "orphaned comment"},
		},
		{
			name:  "post reaction user and post",
			query: "INSERT INTO post_reactions (user_id, post_id, reaction_type) VALUES (?, ?, ?)",
			args:  []any{999, 999, "like"},
		},
		{
			name:  "comment reaction user and comment",
			query: "INSERT INTO comment_reactions (user_id, comment_id, reaction_type) VALUES (?, ?, ?)",
			args:  []any{999, 999, "like"},
		},
		{
			name:  "post category post and category",
			query: "INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			args:  []any{999, 999},
		},
		{
			name:  "session user",
			query: "INSERT INTO sessions (id, user_id, csrf_token, expires_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)",
			args:  []any{"orphaned-session", 999, "csrf-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := db.Exec(tt.query, tt.args...); err == nil {
				t.Fatal("expected insert with missing parent row to fail")
			}
		})
	}
}

// TestConfigureSQLiteConnectionTuning verifies the application opts into
// SQLite settings that prevent normal request bursts from failing immediately.
func TestConfigureSQLiteConnectionTuning(t *testing.T) {
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := configureSQLiteConnection(db); err != nil {
		t.Fatal(err)
	}

	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var busyTimeout int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatal(err)
	}
	if busyTimeout < sqliteBusyTimeoutMilliseconds {
		t.Fatalf("busy_timeout = %d, want at least %d", busyTimeout, sqliteBusyTimeoutMilliseconds)
	}

	stats := db.Stats()
	if stats.MaxOpenConnections != sqliteMaxOpenConnections {
		t.Fatalf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, sqliteMaxOpenConnections)
	}
}

// TestEnsureSessionCSRFColumnUpgradesLegacyTable verifies existing databases can accept new sessions.
func TestEnsureSessionCSRFColumnUpgradesLegacyTable(t *testing.T) {
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	if err := ensureSessionCSRFColumn(db); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO sessions (id, user_id, csrf_token, expires_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)", "session", 1, "csrf"); err != nil {
		t.Fatalf("legacy sessions table did not accept csrf_token after migration: %v", err)
	}
}

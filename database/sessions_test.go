package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetUserIDBySessionIDRejectsExpiredSessionWithPositiveOffset(t *testing.T) {
	db := openSessionTestDB(t)

	if err := CreateUser(db, "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}
	user, err := GetUserByEmail(db, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	sqliteNow := currentSQLiteTimestamp(t, db)
	expiresAt := sqliteNow.Add(-time.Hour).In(time.FixedZone("UTC+03", 3*60*60))
	if err := CreateSession(db, "expired-session", user.ID, "csrf-token", expiresAt); err != nil {
		t.Fatal(err)
	}

	userID, err := GetUserIDBySessionID(db, "expired-session")
	if err != nil {
		t.Fatal(err)
	}
	if userID != 0 {
		t.Fatalf("expired session user id = %d, want 0", userID)
	}
}

func TestCreateSessionStoresExpiryAsSQLiteUTCTimestamp(t *testing.T) {
	db := openSessionTestDB(t)

	if err := CreateUser(db, "user@example.com", "user", "password"); err != nil {
		t.Fatal(err)
	}
	user, err := GetUserByEmail(db, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	expiresAt := time.Date(2026, 6, 9, 17, 30, 15, 123456789, time.FixedZone("UTC+03", 3*60*60))
	if err := CreateSession(db, "session", user.ID, "csrf-token", expiresAt); err != nil {
		t.Fatal(err)
	}

	var stored string
	if err := db.QueryRow("SELECT expires_at FROM sessions WHERE id = ?", "session").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != "2026-06-09 14:30:15" {
		t.Fatalf("stored expiry = %q, want UTC SQLite timestamp", stored)
	}
}

func openSessionTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	if _, err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			csrf_token TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`); err != nil {
		t.Fatal(err)
	}

	return db
}

func currentSQLiteTimestamp(t *testing.T, db *sql.DB) time.Time {
	t.Helper()

	var raw string
	if err := db.QueryRow("SELECT CURRENT_TIMESTAMP").Scan(&raw); err != nil {
		t.Fatal(err)
	}

	sqliteNow, err := time.Parse("2006-01-02 15:04:05", raw)
	if err != nil {
		t.Fatal(err)
	}
	return sqliteNow
}

package database

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens the forum database file and applies the schema idempotently.
func InitDB() (*sql.DB, error) {
	// sql.Open creates a database handle; it does not immediately guarantee
	// that the file can be reached or that the driver can connect.
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		return nil, err
	}

	// sql.Open validates its arguments lazily, so Ping confirms the database is reachable.
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Keep table definitions in schema.sql so tests and the application can
	// initialize databases from the same source.
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		return nil, err
	}

	// The schema uses CREATE TABLE IF NOT EXISTS, so executing it on every
	// startup creates missing tables without dropping existing data.
	_, err = db.Exec(string(schema))
	if err != nil {
		return nil, err
	}

	// Return the shared handle for handlers and database helper functions.
	return db, nil
}

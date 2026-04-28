package database

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens the forum database and applies the schema.
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		return nil, err
	}

	// sql.Open validates its arguments lazily, so Ping confirms the database is reachable.
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return nil, err
	}

	return db, nil
}

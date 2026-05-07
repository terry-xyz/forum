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

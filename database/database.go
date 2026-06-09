package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultDatabasePath           = "forum.db"
	sqliteBusyTimeoutMilliseconds = 5000
	sqliteMaxOpenConnections      = 1
)

// InitDB opens the forum database file and applies the schema idempotently.
func InitDB(databasePath string) (*sql.DB, error) {
	// sql.Open creates a database handle; it does not immediately guarantee
	// that the file can be reached or that the driver can connect.
	db, err := sql.Open("sqlite3", sqliteDataSourceName(databasePath))
	if err != nil {
		return nil, err
	}

	if err := configureSQLiteConnection(db); err != nil {
		db.Close()
		return nil, err
	}

	// sql.Open validates its arguments lazily, so Ping confirms the database is reachable.
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	// Keep table definitions in schema.sql so tests and the application can
	// initialize databases from the same source.
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		db.Close()
		return nil, err
	}

	// The schema uses CREATE TABLE IF NOT EXISTS, so executing it on every
	// startup creates missing tables without dropping existing data.
	_, err = db.Exec(string(schema))
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := ensureSessionCSRFColumn(db); err != nil {
		db.Close()
		return nil, err
	}

	// Return the shared handle for handlers and database helper functions.
	return db, nil
}

func sqliteDataSourceName(databasePath string) string {
	databasePath = strings.TrimSpace(databasePath)
	if databasePath == "" {
		databasePath = defaultDatabasePath
	}

	separator := "?"
	if strings.Contains(databasePath, "?") {
		separator = "&"
	}

	return fmt.Sprintf("%s%s_foreign_keys=on&_busy_timeout=%d&_journal_mode=WAL", databasePath, separator, sqliteBusyTimeoutMilliseconds)
}

func configureSQLiteConnection(db *sql.DB) error {
	db.SetMaxOpenConns(sqliteMaxOpenConnections)
	db.SetMaxIdleConns(sqliteMaxOpenConnections)

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return err
	}
	_, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMilliseconds))
	return err
}

func ensureSessionCSRFColumn(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(sessions)")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == "csrf_token" {
			return rows.Err()
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.Exec("ALTER TABLE sessions ADD COLUMN csrf_token TEXT NOT NULL DEFAULT ''")
	return err
}

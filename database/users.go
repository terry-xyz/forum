package database

import "database/sql"

// CreateUser inserts a new user into the users table.
func CreateUser(db *sql.DB, email string, username string, password string) error {
	query := "INSERT INTO users (email, username, password) VALUES (?, ?, ?)"

	_, err := db.Exec(query, email, username, password)
	if err != nil {
		return err
	}

	return nil
}

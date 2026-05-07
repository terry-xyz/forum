package database

import "database/sql"

// CreateUser inserts a new user into the users table.
func CreateUser(db *sql.DB, email string, username string, password string) error {
	// Email and username uniqueness is enforced by schema constraints, so this
	// insert can return a constraint error for duplicates.
	query := "INSERT INTO users (email, username, password) VALUES (?, ?, ?)"

	// Parameter placeholders keep user-supplied form values out of the SQL text.
	_, err := db.Exec(query, email, username, password)
	if err != nil {
		return err
	}

	// No model is returned because callers either redirect or look up the user
	// later by email.
	return nil
}

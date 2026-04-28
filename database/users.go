package database

import "database/sql"

func CreateUser(db *sql.DB, email string, username string, password string) error {
	query := "INSERT INTO users (email, username, password) VALUES (?, ?, ?)"

	_, err := db.Exec(query, email, username, password)
	if err != nil {
		return err
	}

	return nil
}

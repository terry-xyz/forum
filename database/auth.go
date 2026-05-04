package database

import (
	"database/sql"
	"forum/models"
)

// GetUserByEmail returns the user with the given email, or nil when no user exists.
func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {

	query := "SELECT id, email, username, password FROM users WHERE email = ?"
	row := db.QueryRow(query, email)

	var user models.User

	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID returns the user with the given ID, or nil when no user exists.
func GetUserByID(db *sql.DB, id int) (*models.User, error) {

	query := "SELECT id, email, username, password FROM users WHERE id = ?"
	row := db.QueryRow(query, id)

	var user models.User

	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

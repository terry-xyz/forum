package database

import (
	"database/sql"
	"forum/models"
)

// GetUserByEmail returns the user with the given email, or nil when no user exists.
func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {

	// Look up only the fields needed to authenticate and populate a User model.
	query := "SELECT id, email, username, password FROM users WHERE email = ?"
	row := db.QueryRow(query, email)

	// Scan into a local value first, then return its address only on success.
	var user models.User

	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		// Missing users are not exceptional for login; callers distinguish this
		// with a nil user and nil error.
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID returns the user with the given ID, or nil when no user exists.
func GetUserByID(db *sql.DB, id int) (*models.User, error) {

	// Session validation and author rendering both use the primary-key lookup.
	query := "SELECT id, email, username, password FROM users WHERE id = ?"
	row := db.QueryRow(query, id)

	// Keep the scan shape aligned with the explicit SELECT column order.
	var user models.User

	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		// A missing row means an invalid/stale session or broken author
		// reference, so callers decide which HTTP status to return.
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

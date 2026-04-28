package database

import (
	"database/sql"
	"forum/models"
)

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

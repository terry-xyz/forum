package models

// User represents one row from the users table.
type User struct {
	ID       int
	Email    string
	Username string
	Password string
}
